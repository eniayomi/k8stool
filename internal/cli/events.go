package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	k8s "k8stool/internal/k8s/client"
	"k8stool/internal/k8s/events"
	"k8stool/pkg/utils"

	"github.com/spf13/cobra"
)

func getEventsCmd() *cobra.Command {
	var namespace string
	var allNamespaces bool
	var resourceType string
	var resourceName string
	var component string
	var sortBy string
	var reverse bool
	var since time.Duration
	var watch bool
	var warningsOnly bool

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Get events",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.NewClient()
			if err != nil {
				return err
			}

			// If namespace flag not provided and not all namespaces, use the client's current namespace
			if !allNamespaces && namespace == "" {
				ctx, err := client.GetCurrentContext()
				if err != nil {
					return err
				}
				namespace = ctx.Namespace
			}

			// Create event filter
			filter := &events.EventFilter{
				ResourceKinds: []string{},
				ResourceNames: []string{},
				Components:    []string{},
				SortBy:        events.EventSortOption(sortBy),
			}

			if warningsOnly {
				filter.Types = []events.EventType{events.Warning}
			}

			if resourceType != "" {
				filter.ResourceKinds = append(filter.ResourceKinds, resourceType)
			}

			if resourceName != "" {
				filter.ResourceNames = append(filter.ResourceNames, resourceName)
			}

			if component != "" {
				filter.Components = append(filter.Components, component)
			}

			if since > 0 {
				sinceTime := time.Now().Add(-since)
				filter.Since = &sinceTime
			}

			ctx := context.Background()

			if watch {
				// Watch events
				opts := &events.EventOptions{
					Filter:         filter,
					IncludeManaged: false,
					BufferSize:     100,
				}

				eventChan, err := client.EventService.Watch(ctx, namespace, opts)
				if err != nil {
					return err
				}

				for event := range eventChan {
					printEvent(&event)
				}

				return nil
			}

			// List events
			eventList, err := client.EventService.List(ctx, namespace, filter)
			if err != nil {
				return err
			}

			return printEvents(eventList.Items)
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List events across all namespaces")
	cmd.Flags().StringVar(&resourceType, "resource-type", "", "Filter events by resource type")
	cmd.Flags().StringVar(&resourceName, "resource-name", "", "Filter events by resource name")
	cmd.Flags().StringVar(&component, "component", "", "Filter events by component")
	cmd.Flags().StringVar(&sortBy, "sort", string(events.SortByTime), "Sort by (time, count, type, resource)")
	cmd.Flags().BoolVar(&reverse, "reverse", false, "Reverse sort order")
	cmd.Flags().DurationVar(&since, "since", 0, "Show events newer than a relative duration")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch events")
	cmd.Flags().BoolVar(&warningsOnly, "warnings", false, "Show only warning events")

	return cmd
}

func printEvents(events []events.Event) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "LAST SEEN\tTYPE\tREASON\tOBJECT\tMESSAGE")

	for _, e := range events {
		age := utils.FormatDuration(time.Since(e.LastTimestamp))
		object := fmt.Sprintf("%s/%s", e.ResourceKind, e.ResourceName)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			age,
			utils.ColorizeEventType(string(e.Type)),
			e.Reason,
			object,
			e.Message)
	}

	return nil
}

func printEvent(e *events.Event) {
	age := utils.FormatDuration(time.Since(e.LastTimestamp))
	object := fmt.Sprintf("%s/%s", e.ResourceKind, e.ResourceName)
	fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
		age,
		utils.ColorizeEventType(string(e.Type)),
		e.Reason,
		object,
		e.Message)
}
