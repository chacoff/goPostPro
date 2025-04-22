/*
 * File:    klusters.go
 * Date:    March 09, 2025
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Using K-means clusters all measures in order to re-assign the Pass number
 *
 */

package klusters

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"goPostPro/global"
	"log"
	"sort"
)

// ReClusterPasses will re-assign the pass by using kmeans over the timestamps of the measures
func ReClusterPasses(beginTS string, endTS string, beamID uint32) {
	var firstCoordinates clusters.Coordinates
	var lastCoordinates clusters.Coordinates
	var k int = global.Cluster.K
	var delta float64 = global.Cluster.Delta

	timestamps, errorGetData := getData(beginTS, endTS, beamID)

	if errorGetData != nil {
		log.Println("[CLUSTER] Error getting data from DB to start clustering")
		return
	}
	unixTimes, timestampMap := returnUnixTS(timestamps)

	// Clustering, convert 1D data to clusters.Observations
	var observations clusters.Observations

	for _, value := range unixTimes {
		observations = append(observations, clusters.Coordinates{value})
	}

	//km := kmeans.New()
	km, _ := kmeans.NewWithOptions(delta, nil)

	clustersPasses, err := km.Partition(observations, k)
	if err != nil {
		log.Println("[CLUSTER] Error partitioning data: ", err)
	}

	sort.Slice(clustersPasses, func(i, j int) bool {
		// Skip empty clusters
		if len(clustersPasses[i].Observations) == 0 {
			return false
		}
		if len(clustersPasses[j].Observations) == 0 {
			return true
		}

		// Get first observation from each cluster
		firstI := clustersPasses[i].Observations[0].(clusters.Coordinates)[0]
		firstJ := clustersPasses[j].Observations[0].(clusters.Coordinates)[0]

		// Sort by earliest timestamp (ascending order)
		return firstI < firstJ
	})

	// Output results
	for i, c := range clustersPasses {

		if len(c.Observations) > 0 {
			firstObs := c.Observations[0]
			lastObs := c.Observations[len(c.Observations)-1]

			firstCoordinates = firstObs.(clusters.Coordinates)
			lastCoordinates = lastObs.(clusters.Coordinates)
		}

		pass := fmt.Sprintf("Pass %d", i+1)
		errorUpdate := updatePass(timestampMap[firstCoordinates[0]], timestampMap[lastCoordinates[0]], pass, beamID)
		if errorUpdate != nil {
			return
		}
	}
}
