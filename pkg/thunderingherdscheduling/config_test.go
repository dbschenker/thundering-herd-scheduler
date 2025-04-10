package thunderingherdscheduling

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"testing"
)

func TestParseArguments(t *testing.T) {
	testcases := []struct {
		name        string
		input       string
		expected    *ThunderingHerdSchedulingArgs
		errExpected bool
		errMsg      string
	}{
		{
			name:        "empty",
			input:       `{}`,
			expected:    &ThunderingHerdSchedulingArgs{},
			errExpected: false,
		},
		{
			name:  "parallelStartingPodsPerNode",
			input: `{"parallelStartingPodsPerNode": 5}`,
			expected: &ThunderingHerdSchedulingArgs{
				ParallelStartingPodsPerNode: ptr.To(5),
			},
			errExpected: false,
		},
		{
			name:  "parallelStartingPodsPerCore",
			input: `{"parallelStartingPodsPerCore": 10}`,
			expected: &ThunderingHerdSchedulingArgs{
				ParallelStartingPodsPerCore: ptr.To(10),
			},
			errExpected: false,
		},
		{
			name:        "both",
			input:       `{"parallelStartingPodsPerCore": 10, "parallelStartingPodsPerNode": 5}`,
			expected:    nil,
			errExpected: true,
			errMsg:      "cannot specify parallelStartingPodsPerNode and parallelStartingPodsPerCore at the same time",
		},
		{
			name:        "malformed",
			input:       `wrong json`,
			expected:    nil,
			errExpected: true,
			errMsg:      "invalid character 'w' looking for beginning of value",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			unk := runtime.Unknown{Raw: []byte(tc.input)}
			out, err := ParseArguments(&unk)
			assert.Equal(t, tc.expected, out)
			//assert.Equal(t, tc.err, err)
			if tc.errExpected {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
