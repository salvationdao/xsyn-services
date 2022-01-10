package deposit

import "fmt"

func DisplayAddress(in string) string {
	return fmt.Sprintf("%s...%s", in[:5], in[len(in)-5:])
}
