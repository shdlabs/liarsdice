package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/ardanlabs/liarsdice/foundation/vault"
)

// vaultCmd represents the vault command
var vaultInitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize/unseal a vault instance",
	Long:  `Used to initialize and unseal a Vault instance in an idempotent manner.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultConfig, err := getVaultConfig(cmd)
		if err != nil {
			return err
		}

		credentialsFile, err := cmd.Flags().GetString(credentialFile)
		if err != nil {
			return err
		}

		return vaultInit(credentialsFile, vaultConfig)
	},
}

const defaultCredentialFile = "/tmp/credentials.json"

func init() {
	vaultCmd.AddCommand(vaultInitCmd)
	vaultInitCmd.Flags().StringP(credentialFile, shortName[credentialFile], defaultCredentialFile, "Credential file generated by init")
}

func vaultInit(credentialFile string, vaultConfig vault.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	vaultSrv, err := vault.New(vaultConfig)
	if err != nil {
		return fmt.Errorf("constructing vault: %w", err)
	}

	initResponse, err := checkIfCredFileExists(credentialFile)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			log.Println("credential file doesn't exist, initializing vault")

			initResponse, err = vaultSrv.SystemInit(ctx, 1, 1)
			if err != nil {
				if errors.Is(err, vault.ErrAlreadyInitialized) {
					return fmt.Errorf("vault is already initialized but we don't have the credentials file")
				}
				return fmt.Errorf("unable to initialize Vault instance: %w", err)
			}

			b, err := json.Marshal(initResponse)
			if err != nil {
				return errors.New("unable to marshal")
			}

			if err := os.WriteFile(credentialFile, b, 0644); err != nil {
				return fmt.Errorf("unable to write %s file: %w", credentialFile, err)
			}

		default:
			return fmt.Errorf("unable to read credentials file: %w", err)
		}
	}

	log.Printf("rootToken: %s", initResponse.RootToken)

	// =============================================================================================================

	log.Println("Unsealing vault")
	err = vaultSrv.Unseal(ctx, initResponse.KeysB64[0])
	if err != nil {
		if errors.Is(err, vault.ErrBadRequest) {
			return fmt.Errorf("vault is not initialized. Check for old credentials file: %s", credentialFile)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	// =============================================================================================================

	log.Println("Mounting path in vault")

	vaultSrv.SetToken(initResponse.RootToken)
	if err := vaultSrv.Mount(ctx); err != nil {
		if errors.Is(err, vault.ErrPathInUse) {
			return fmt.Errorf("unable to mount path: %w", err)
		}
		return fmt.Errorf("error unsealing vault: %w", err)
	}

	// =============================================================================================================

	log.Println("Creating sales-api policy")

	err = vaultSrv.CreatePolicy(ctx, "sales-api", "secret/data/*", []string{"read", "create", "update"})
	if err != nil {
		return fmt.Errorf("unable to create policy: %w", err)
	}

	// =============================================================================================================

	log.Printf("Generating sales-api token: %s", vaultConfig.Token)

	// First let's check if it exists already.
	err = vaultSrv.CheckToken(ctx, vaultConfig.Token)
	if err == nil {
		log.Printf("token already exists: %s", vaultConfig.Token)
		return nil
	}

	// We don't currently save the token because we're always going to specify it.
	err = vaultSrv.CreateToken(ctx, vaultConfig.Token, []string{"sales-api"}, "Sales API")
	if err != nil {
		return fmt.Errorf("unable to create token: %w", err)
	}

	return nil
}

// =============================================================================

func checkIfCredFileExists(credentialFile string) (vault.SystemInitResponse, error) {
	if _, err := os.Stat(credentialFile); err != nil {
		return vault.SystemInitResponse{}, err
	}

	data, err := os.ReadFile(credentialFile)
	if err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("reading %s file: %s", credentialFile, err)
	}

	var response vault.SystemInitResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return vault.SystemInitResponse{}, fmt.Errorf("unmarshalling json: %s", err)
	}

	return response, nil
}
