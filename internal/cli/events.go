package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"k8stool/internal/k8s"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getEventsCmd() *cobra.Command {
	var namespace string

	cmd := &cobra.Command{
		Use:     "events [pod-name]",
		Aliases: []string{"ev"},
		Short:   "Show events for a pod",
		Long: `Show events for a pod.
Example: k8stool get events nginx-pod
         k8stool get ev nginx-pod`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			podName := args[0]
			events, err := client.GetPodEvents(namespace, podName)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
			fmt.Fprintln(w, "LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE")

			for _, event := range events {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					utils.FormatDuration(event.LastSeen),
					utils.ColorizeEventType(event.Type),
					event.Reason,
					event.Object,
					event.Message)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	return cmd
}
