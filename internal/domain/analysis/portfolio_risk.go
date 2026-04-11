package analysis

import "fmt"

const (
	singlePositionConcentrationThreshold = 0.20
	top3ConcentrationThreshold           = 0.60
)

type RiskAnalyzer struct{}

func NewRiskAnalyzer() *RiskAnalyzer {
	return &RiskAnalyzer{}
}

type PortfolioRisk struct {
	PortfolioName     string
	BaseCurrency      string
	PositionCount     int
	LargestPosition   string
	LargestWeight     float64
	Top3Concentration float64
	Observations      []string
}

func (a *RiskAnalyzer) Analyze(in PortfolioAnalysis) (PortfolioRisk, error) {
	out := PortfolioRisk{
		PortfolioName: in.PortfolioName,
		BaseCurrency:  in.BaseCurrency,
		PositionCount: len(in.AnalyzedPositions),
		Observations:  make([]string, 0),
	}

	if len(in.AnalyzedPositions) == 0 {
		out.Observations = append(out.Observations, "No positions in portfolio")
		return out, nil
	}

	largest := in.AnalyzedPositions[0]
	out.LargestPosition = largest.Symbol
	out.LargestWeight = largest.Weight

	topN := 3
	if len(in.AnalyzedPositions) < topN {
		topN = len(in.AnalyzedPositions)
	}

	for i := 0; i < topN; i++ {
		out.Top3Concentration += in.AnalyzedPositions[i].Weight
	}

	if out.LargestWeight >= singlePositionConcentrationThreshold {
		out.Observations = append(
			out.Observations,
			fmt.Sprintf("Single-position concentration is elevated: %s is %.2f%% of the portfolio", out.LargestPosition, out.LargestWeight*100),
		)
	}

	if out.Top3Concentration >= top3ConcentrationThreshold {
		out.Observations = append(
			out.Observations,
			fmt.Sprintf("Top 3 concentration is high at %.2f%%", out.Top3Concentration*100),
		)
	}

	if len(out.Observations) == 0 {
		out.Observations = append(out.Observations, "No obvious concentration issues detected")
	}

	return out, nil
}
