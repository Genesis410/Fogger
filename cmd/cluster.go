package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster <cluster-id>",
	Short: "View all domains and evidence connected to an operator/campaign",
	Long: `Cluster shows all domains and evidence connected to a specific 
operator or campaign cluster.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterID := args[0]
		graph, _ := cmd.Flags().GetBool("graph")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		since, _ := cmd.Flags().GetString("since")

		fmt.Printf("Cluster ID: %s\n", clusterID)
		fmt.Printf("Graph view: %t\n", graph)
		fmt.Printf("JSON output: %t\n", jsonOutput)
		fmt.Printf("Since: %s\n", since)
		
		// In a real implementation, this would fetch cluster data
		fmt.Println("Cluster data would be displayed here...")
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	// Add flags for the cluster command
	clusterCmd.Flags().Bool("graph", false, "ASCII graph visualization")
	clusterCmd.Flags().Bool("json", false, "Output JSON")
	clusterCmd.Flags().String("since", "", "Time filter (e.g., 30d)")
}