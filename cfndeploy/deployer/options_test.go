package deployer

import (
	"testing"
)

func TestStackParamsAsAWSParams(t *testing.T) {
	params := StackParams(map[string]string{
		"paramOne": "valueOne",
		"paramTwo": "valueTwo",
	})

	awsparams := params.AWSParams()

	if len(awsparams) != 2 {
		t.Errorf("Incorrect length. Want %d, got %d", 2, len(awsparams))
	}

	for _, a := range awsparams {
		if params[*a.ParameterKey] != *a.ParameterValue {
			t.Errorf("Param value mismatch. Want %s, got %s", params[*a.ParameterKey], *a.ParameterValue)
		}
	}
}
