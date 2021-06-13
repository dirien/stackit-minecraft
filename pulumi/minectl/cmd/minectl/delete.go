package minectl

import (
	"github.com/minectl/pgk/provisioner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {

	minectlCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("filename", "f", "", "that contains the configuration for minectl")

}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an Minecraft Server.",
	Long: `Delete an an Minecraft Server
	`,
	Example: `  mincetl delete  \
    --filename server.yaml`,
	RunE:          runDelete,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func runDelete(cmd *cobra.Command, _ []string) error {

	filename, err := cmd.Flags().GetString("filename")
	if err != nil {
		return errors.Wrap(err, "failed to get 'filename' value.")
	}
	do := provisioner.NewProvisioner(filename)
	_, err = do.UpdateServer()
	return err
}
