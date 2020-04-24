// The cmd package provides the main entry point into the vssh program
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
	"path/filepath"
)

var server string
var token string
var role string
var mount string
var persist bool
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
		main(args)
	},
}

// main is executed by the root command and is the main entry point to the program
func main(args []string) {
	publicKeyPath, pubKeyBytes, err := ssh.GetPublicKey(viper.GetString("identity"))
	if err != nil {
		errorThenExit("Error fetching public key", err)
	}

	// Check if a cert exists and is still valid
	certPath := ssh.GetPublicKeyCertPath(publicKeyPath)
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		cert, err := ssh.GetCertificate(certPath)
		if err != nil {
			errorThenExit("Error reading certificate at " + certPath, err)
		}

		if ssh.IsCertificateValid(cert) {
			runSSH(args) // No need to continue further since the cert is still valid
		}
	}

	// Must have a role specified at this point
	if viper.GetString("role") == "" {
		fmt.Println("Please specify a role to sign with")
		os.Exit(1)
	}

	vaultClient, err := client.NewDefaultClient()
	if err != nil {
		errorThenExit("Error trying to load Vault client configuration", err)
	}

	if err := vaultClient.SetConfigValues(viper.GetString("server"), viper.GetString("token")); err != nil {
		errorThenExit("Error setting Vault server or token: ", err)
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

	if !vaultClient.Authenticated() {
		login(vaultClient)
	}

	signedKey, err := vaultClient.SignPubKey(viper.GetString("mount"), viper.GetString("role"), pubKeyBytes)
	if err != nil {
		errorThenExit("Error signing public key", err)
	}

	if err := ioutil.WriteFile(certPath, []byte(signedKey), 0644); err != nil {
		errorThenExit("Error writing public key certificate", err)
	}

	fmt.Println("Wrote certificate to ", certPath)
	runSSH(args)
}

// login performs the process of requesting credentials from the end-user and using them to perform a login against the
// given VaultClient instance.
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

	if viper.GetBool("persist") {
		home, err := homedir.Dir()
		if err != nil {
			errorThenExit("Error getting user home directory", err)
		}

		tokenPath := filepath.Join(home, ".vault-token")
		if err := ioutil.WriteFile(tokenPath, []byte(vaultClient.Token()), 0644); err != nil {
			errorThenExit("Error persising token to ~/.vault-token", err)
		}
	}
}

// runSSH creates and executes the ssh command using the given arguments
func runSSH(args []string) {
	cmd := ssh.NewSSHCommand(args)
	if err := cmd.Run(); err != nil {
		errorThenExit("Error running ssh command", err)
	}
	os.Exit(0)
}

// errorThenExit is a small wrapper for reporting and error and existing with a non-zero exit code
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

// init is where flags and configuration data is setup
func init() {
	// Load config file
	cobra.OnInitialize(initConfig)

	// Vault variables
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "address of vault server (default: $VAULT_ADDR)")
	err := viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))

	rootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "vault token to use for authentication (default: $VAULT_TOKEN)")
	err = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))

	rootCmd.PersistentFlags().StringVarP(&role, "role", "r", "", "vault role account to sign with")
	err = viper.BindPFlag("role", rootCmd.PersistentFlags().Lookup("role"))

	rootCmd.PersistentFlags().StringVarP(&mount, "mount", "m", "", "mount path for ssh backend (default: ssh)")
	err = viper.BindPFlag("mount", rootCmd.PersistentFlags().Lookup("mount"))

	rootCmd.PersistentFlags().BoolVarP(&persist, "persist", "p", false, "persist obtained tokens to ~/.vault-token")
	err = viper.BindPFlag("persist", rootCmd.PersistentFlags().Lookup("persist"))

	// SSH variables
	rootCmd.PersistentFlags().StringVarP(&identity, "identity", "i", "", "ssh key-pair to sign and use (default: $HOME/.ssh/id_rsa)")
	err = viper.BindPFlag("identity", rootCmd.PersistentFlags().Lookup("identity"))

	// Config variables
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.vssh)")

	if err != nil {
		errorThenExit("Error binding to flags", err)
	}
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
			errorThenExit("Error getting user home directory", err)
		}

		// Default to $HOME/.vssh
		viper.SetConfigFile(filepath.Join(home, ".vssh"))
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("VSSH")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
