package cmd

import (
	"fmt"
	"github.com/jmgilman/vssh/client"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var server string
var token string
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
	// Attempt to create a Client using default config parameters
	client, err := client.NewDefaultClient()
	if err != nil {
		fmt.Println("Error trying to load Vault client configuration: ", err)
		os.Exit(1)
	}

	// Verify the vault is in a usable state
	status, err := client.Available()
	if err != nil {
		fmt.Println("Error trying to check vault status: ", err)
	}

	if !status {
		fmt.Println("The vault is either sealed or not initialized - cannot continue")
		os.Exit(1)
	}

	// Check if client is authenticated - if not, offer authentication options
	// TODO(jmgilman): Add authentication and login steps
	if client.Authenticated() {
		fmt.Println("Authenticated!")
	} else {
		fmt.Println("Not authenticated!")
	}
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
