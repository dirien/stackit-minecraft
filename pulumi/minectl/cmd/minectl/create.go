package minectl

import (
	"github.com/minectl/pgk/common"
	"github.com/minectl/pgk/provisioner"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {

	minectlCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("filename", "f", "", "that contains the configuration for minectl")

}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an Minecraft Server.",
	Long: `Create an an Minecraft Server
	`,
	Example: `  mincetl create  \
    --filename server-do.yaml`,
	RunE:          runCreate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runCreate(cmd *cobra.Command, _ []string) error {

	filename, err := cmd.Flags().GetString("filename")
	if err != nil {
		return errors.Wrap(err, "failed to get 'filename' value.")
	}
	do := provisioner.NewProvisioner(filename)
	res, err := do.CreateServer()
	if err != nil {
		log.Fatal(err)
	}
	common.PrintMixedGreen("Minecraft Server IP: %s\n", res.PublicIP)
	return err
}
