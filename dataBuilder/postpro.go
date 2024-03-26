/*
 * File:    postpro.go
 * Date:    March 04, 2024
 * Author:  J.
 * Email:   jaime.gomez@usach.cl
 * Project: goPostPro
 * Description:
 *   Gathers data from thermal cameras at Train2 and cross-match with timestamps coming from MES to
 *	 to outcome post processes data.
 *
 * Gonum install:
 *   go get gonum.org/v1/gonum@latest
 *   go test gonum.org/v1/gonum
 */

package dataBuilder

import (
	"fmt"
	"gonum.org/v1/gonum/stat"
	"math/rand/v2"
)

func Calculus() {

	var (
		xs      = make([]float64, 100)
		ys      = make([]float64, 100)
		weights []float64
	)

	line := func(x float64) float64 {
		return 1 + 3*x
	}

	for i := range xs {
		xs[i] = float64(i)
		ys[i] = line(xs[i]) + 0.1*rand.NormFloat64()
	}

	alpha, beta := stat.LinearRegression(xs, ys, weights, false) // Do not force the regression line to pass through the origin.
	r2 := stat.RSquared(xs, ys, weights, alpha, beta)

	fmt.Printf("Estimated offset is: %.6f\n", alpha)
	fmt.Printf("Estimated slope is:  %.6f\n", beta)
	fmt.Printf("R^2: %.6f\n", r2)

}
