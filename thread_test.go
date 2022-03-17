package models

import (
	"testing"
)

func TestThread(t *testing.T) {
	raw := `{"length":2000,"content":[{"type":"paragraph","content":[{"style":"normal","content":"这是一个段落\n"},{"style":"bold","content":"支持粗体、"},{"style":"underline","content":"下划线、"},{"style":"italic","content":"斜体、"},{"style":"del","content":"删除线、"},{"style":"bold,underline","content":"混合。"}]},{"type":"image","url":"https://pic1.zhimg.com/v2-064053037ffdff311bff33d1b1184db8_1440w.jpg","desc":"这里是图片描述"},{"type":"reference","content":"这是一段引用"},{"type":"header","level":1,"content":"这是一级标题"},{"type":"ul","items":["这是一个无序列表1","这是一个无序列表2"]},{"type":"ol","items":["这是一个有序列表1","这是一个有序列表2"]}]}`

	err := Connect(getDcs())
	if err != nil {
		t.Error("failed to connect to database.")
		t.Fatal(err)
	}

	_, err = NewPost("test_thread", raw, 1)
	if err != nil {
		t.Error(err)
	}
}
