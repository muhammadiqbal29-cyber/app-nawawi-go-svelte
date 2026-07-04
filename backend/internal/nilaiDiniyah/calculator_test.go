package nilaiDiniyah

import (
	"testing"
)

func TestCalculateNaPTS_Remidi(t *testing.T) {
	remidi := 70
	na := calculateNaPTS(50, &remidi)
	if na != KKM {
		t.Errorf("Expected %d, got %d", KKM, na)
	}

	naMurniTinggi := calculateNaPTS(80, nil)
	if naMurniTinggi != 80 {
		t.Errorf("Expected 80, got %d", naMurniTinggi)
	}

	var zeroRemidi = 0
	naMurniRendah := calculateNaPTS(50, &zeroRemidi)
	if naMurniRendah != 50 {
		t.Errorf("Expected 50, got %d", naMurniRendah)
	}
}

func TestClampScore_Negative(t *testing.T) {
	if clampScore(-10) != 0 {
		t.Error("Expected 0")
	}
	if clampScore(110) != 100 {
		t.Error("Expected 100")
	}
}

func TestEncodeUhScores_Error(t *testing.T) {
	if encodeUhScores(nil) != "[]" {
		t.Error("Expected []")
	}
}
