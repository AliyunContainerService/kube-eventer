package dingtalk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/olekukonko/tablewriter"
	//"os"
	"time"
)

func TestGetLevel(t *testing.T) {
	warning := getLevel(v1.EventTypeWarning)
	normal := getLevel(v1.EventTypeNormal)
	none := getLevel("")
	assert.True(t, warning > normal)
	assert.True(t, warning == WARNING)
	assert.True(t, normal == NORMAL)
	assert.True(t, 0 == none)
}

func TestCreateMsgFromEvent(t *testing.T) {
	now := time.Now()
	labels := make([]string, 1)
	labels[0]="abcd"
	event := &v1.Event{
		Message:        "some thing wrong",
		Count:          251,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),		
	}

	msg := createMsgFromEvent(labels,MARKDOWN_MSG_TYPE, event)
	t.Log(msg.Text)
	assert.True(t, msg != nil)
}

//
//func TestTableWriter(t *testing.T) {
//
//	data := [][]string{
//		[]string{"A", "The Good", "500"},
//		[]string{"B", "The Very very Bad Man", "288"},
//		[]string{"C", "The Ugly", "120"},
//		[]string{"D", "The Gopher", "800"},
//	}
//
//	table := tablewriter.NewWriter(os.Stdout)
//	table.SetHeader([]string{"Sign"})
//
//	for _, v := range data {
//		table.Append(v)
//	}
//	table.Render() // Send output
//}
