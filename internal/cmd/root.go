package cmd

import (
	"fmt"
	"github.com/jmgilman/vssh/auth"
	"github.com/jmgilman/vssh/client"
	"github.com/jmgilman/vssh/internal/ui"
	"github.com/jmgilman/vssh/ssh"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

var server string
var token string
var role string
var mount string
var identity string

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vssh [ssh host] [flags] -- [ssh-flags]",
	Short: "A small wrapper for authenticating with SSH keys from Hashicorp Vault",
	Long: `This wrapper automatically handles the process of fetching a signed public key certificate from a Hashicorp 
Vault instance and using it to authenticate against a given host. It uses the default Vault environment variables for
setting the Vault server address and token, but all configuration details can be optionally provided via flags, a 
custom config file, or the default configuration file at ~/.vssh. If no token is provided, the wrapper will 
automatically prompt to authenticate against Vault and obtain a new token via any configured authentication method.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		main()
		//TODO(jmgilman) Implement forwarding SSH
		//ui.CallSSH(args)
	},
}

// main is executed by the root command and is the main entry point to the program
func main() {
	// See if the public key certificate exists and is signed
	publicKeyPath, pubKeyBytes, err := ssh.GetPublicKey(identity)
	if err != nil {
		errorThenExit("Error fetching public key", err)
	}
	certPath := ssh.GetPublicKeyCertPath(publicKeyPath)
	valid, err := ssh.IsCertificateValid(certPath)
	if err != nil {
		errorThenExit("Error reading certificate at " + certPath, err)
	}
	if valid {
		fmt.Println("Already signed!")
		os.Exit(0)
	}

	// Attempt to create a Client using default config parameters
	vaultClient, err := client.NewDefaultClient()
	if err != nil {
		errorThenExit("Error trying to load Vault client configuration", err)
	}

	// Verify the vault is in a usable state
	status, err := vaultClient.Available()
	if err != nil {
		errorThenExit("Error trying to check vault status", err)
	}

	if !status {
		fmt.Println("The vault is either sealed or not initialized - cannot continue")
		os.Exit(1)
	}

	// Check if client is authenticated - if not, attempt to perform a login with the client
	if !vaultClient.Authenticated() {
		login(vaultClient)
	}

	signedKey, err := vaultClient.SignPubKey(mount, role, pubKeyBytes)
	if err != nil {
		errorThenExit("Error signing public key", err)
	}

	if err := ioutil.WriteFile(certPath, []byte(signedKey), 0644); err != nil {
		errorThenExit("Error writing public key certificate", err)
	}

	fmt.Println("Wrote certificate to ", certPath)
}

func login(vaultClient *client.VaultClient) {
	// Ask which authentication type they would like to use
	prompt := ui.NewSelectPrompt("Please choose an authentication method:", auth.GetAuthNames())
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Println("Error getting authentication method:", err)
		os.Exit(1)
	}

	// Collect authentication details for the selected method
	authType := auth.Types[result]()
	details, err := ui.GetAuthDetails(authType, ui.NewPrompt)

	// Login with the collected details
	if err := vaultClient.Login(authType, details); err != nil {
		fmt.Println("Error logging in:", err)
		os.Exit(1)
	}

	fmt.Println("Authentication successful!")
}

func errorThenExit(message string, err error) {
	fmt.Println(message, ":", err)
	os.Exit(1)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// Vault variables
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "address of vault server")
	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "vault token to use for authentication")
	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "vault role account to sign with")
	rootCmd.PersistentFlags().StringVarP(&mount, "mount", "m", "", "mount path for ssh backend")

	// SSH variables
	rootCmd.PersistentFlags().StringVarP(&identity, "identity", "i", "", "ssh key-pair to sign and use (defaults to $HOME/.ssh/id_rsa)")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vssh)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vssh" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vssh")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
