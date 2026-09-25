// Package performance provides deterministic performance observations for compiler stages.
//
// Stage implementations supply an operation counter and may be used from regular
// tests or Go benchmarks. Budget checks intentionally use operation and allocation
// counts; wall-clock time remains a diagnostic benchmark signal only.
package performance
