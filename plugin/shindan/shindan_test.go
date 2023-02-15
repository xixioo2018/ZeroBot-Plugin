package shindan

import (
	"fmt"
	"github.com/FloatTech/AnimeAPI/shindanmaker"
	"testing"
)

func TestName(t *testing.T) {
	txt, err := shindanmaker.Shindanmaker(162207, "妖诱")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(txt)
}
