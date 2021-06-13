package minectl

import (
	"github.com/minectl/pgk/provisioner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {

	minectlCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringP("filename", "f", "", "that contains the configuration for minectl")

}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an Minecraft Server.",
	Long: `Update an an Minecraft Server
	`,
	Example: `  mincetl update  \
    --filename server.yaml`,
	RunE:          runUpdate,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runUpdate(cmd *cobra.Command, _ []string) error {

	filename, err := cmd.Flags().GetString("filename")
	if err != nil {
		return errors.Wrap(err, "failed to get 'filename' value.")
	}
	do := provisioner.NewProvisioner(filename)
	_, err = do.UpdateServer()
	return err
}
