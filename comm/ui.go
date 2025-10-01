package comm

import (
	"fmt"
	"os"
	"slices"
)

var yesList = []string{
	"y",
	"Y",
	"yes",
	"Yes",
	"YES",
}

var noList = []string{
	"n",
	"N",
	"no",
	"No",
	"NO",
}

func UI(ask string, defaultAnswer string, runner func(bool)) {
	answer := defaultAnswer
	for {
		OutputUI(os.Stdout, ask)
		c, err := fmt.Scanln(&answer)
		if err != nil {
			if err.Error() != "unexpected newline" {
				OutputError("输入发生错误, 请重新输入", err.Error())
				continue
			}
		}
		if c == 0 || answer == "" {
			answer = defaultAnswer
		}
		if !slices.Contains(yesList, answer) && !slices.Contains(noList, answer) {
			OutputError("输入不符合期望, 请重新输入")
			continue
		}
		break
	}
	if slices.Contains(yesList, answer) {
		runner(true)
	} else {
		runner(false)
	}
}
