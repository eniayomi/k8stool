package cli

import (
	"fmt"
	k8s "k8stool/internal/k8s/client"
	"k8stool/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

func getEventsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var selector string
	var watch bool
	var resourceType string
	var resourceName string

	cmd := &cobra.Command{
		Use:   "events [TYPE] [NAME]",
		Short: "Get events for a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			if len(args) >= 2 {
				resourceType = args[0]
				resourceName = args[1]
			}

			if namespace == "" {
				namespace = client.GetNamespace()
			}

			details, err := client.GetDetails(namespace, resourceType, resourceName)
			if err != nil {
				return err
			}

			// Print events
			fmt.Printf("LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE\n")
			for _, e := range details.Events {
				fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
					utils.FormatDuration(time.Since(e.LastSeen)),
					e.Type,
					e.Reason,
					e.Object,
					e.Message,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List events from all namespaces")
	cmd.Flags().StringVarP(&selector, "selector", "l", "", "Selector (label query) to filter on")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")

	return cmd
}
