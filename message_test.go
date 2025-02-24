// Copyright (c) 2017 Ernest Micklei
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package proto

import (
	"testing"
)

func TestMessage(t *testing.T) {
	proto := `
		message   Out   {
		// identifier
		string   id  = 1;
		// size
		int64   size = 2;

		oneof foo {
			string     name        = 4;
			SubMessage sub_message = 9;
		}
		message  Inner {   // Level 2
   			int64  ival = 1;
  		}
		map<string, testdata.SubDefaults> proto2_value  =  13;
		option  (my_option).a  =  true;
	}`
	p := newParserOn(proto)
	p.next() // consume first token
	m := new(Message)
	err := m.parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := m.Name, "Out"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := len(m.Elements), 6; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[0].(*NormalField).Position.String(), "<input>:4:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[0].(*NormalField).Comment.Position.String(), "<input>:3:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[3].(*Message).Position.String(), "<input>:12:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[3].(*Message).Elements[0].(*NormalField).Position.Line, 13; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	checkParent(m.Elements[5].(*Option), t)
}

func TestRepeatedGroupInMessage(t *testing.T) {
	src := `message SearchResponse {
		repeated group Result = 1 {
		  required string url = 2;
		  optional string title = 3;
		  repeated string snippets = 4;
		}
	  }`
	p := newParserOn(src)
	p.next() // consume first token
	m := new(Message)
	err := m.parse(p)
	if err != nil {
		t.Error(err)
	}
	if got, want := len(m.Elements), 1; got != want {
		t.Logf("%#v", m.Elements)
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	g := m.Elements[0].(*Group)
	if got, want := len(g.Elements), 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := g.Repeated, true; got != want {
		t.Fatalf("got Repeated [%v] want [%v]", got, want)
	}

}

func TestRequiredGroupInMessage(t *testing.T) {
	src := `message SearchResponse {
		required group Result = 1 {
		  required string url = 2;
		  optional string title = 3;
		  repeated string snippets = 4;
		}
	  }`
	p := newParserOn(src)
	p.next() // consume first token
	m := new(Message)
	err := m.parse(p)
	if err != nil {
		t.Error(err)
	}
	if got, want := len(m.Elements), 1; got != want {
		t.Logf("%#v", m.Elements)
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	g := m.Elements[0].(*Group)
	if got, want := len(g.Elements), 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := g.Required, true; got != want {
		t.Fatalf("got Required [%v] want [%v]", got, want)
	}

}

func TestSingleQuotedReservedNames(t *testing.T) {
	src := `message Channel {
		reserved '', 'things', "";
	  }`
	p := newParserOn(src)
	p.next() // consume first token
	m := new(Message)
	err := m.parse(p)
	if err != nil {
		t.Error(err)
	}
	r := m.Elements[0].(*Reserved)
	if got, want := r.FieldNames[0], ""; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := r.FieldNames[1], "things"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := r.FieldNames[2], ""; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
}

func TestMessageInlineCommentBeforeBody(t *testing.T) {
	src := `message BarMessage // BarMessage
	  // with another line
	  {
		  name string = 1;
	  } 
	`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
	nestedComment := msg.Elements[0].(*Comment)
	if nestedComment == nil {
		t.Fatal("expected comment present")
	}
	if got, want := len(nestedComment.Lines), 2; got != want {
		t.Errorf("got %d want %d lines", got, want)
	}
}

func TestMessageWithMessage(t *testing.T) {
	src := `message message {
		string message = 1;
	}
	`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
	if got, want := msg.Name, "message"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
	if got, want := len(msg.Elements), 1; got != want {
		t.Errorf("got %d want %d elements", got, want)
	}
	f := msg.Elements[0].(*NormalField)
	if got, want := f.Name, "message"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

func TestIssue143_Key(t *testing.T) {
	src := `message Msg {
  option (option_name) = { [key]: value_name };
}`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
	name := msg.Elements[0].(*Option).AggregatedConstants[0].Name
	value := msg.Elements[0].(*Option).AggregatedConstants[0].Literal.Source
	if got, want := name, "key"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
	if got, want := value, "value_name"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
}
func TestIssue143_KeyDot(t *testing.T) {
	src := `message Msg {
  option (option_name) = { [key.dot]: value_name };
}`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
	name := msg.Elements[0].(*Option).AggregatedConstants[0].Name
	value := msg.Elements[0].(*Option).AggregatedConstants[0].Literal.Source
	if got, want := name, "key.dot"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
	if got, want := value, "value_name"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
}
func TestIssue143_Keyword(t *testing.T) {
	src := `message Msg {
  option (option_name) = { [option.message]: repeated }; 
}`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
	name := msg.Elements[0].(*Option).AggregatedConstants[0].Name
	value := msg.Elements[0].(*Option).AggregatedConstants[0].Literal.Source
	if got, want := name, "option.message"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
	if got, want := value, "repeated"; got != want {
		t.Errorf("got %s want %s", got, want)
	}
}
func TestCommentsInFieldOptionsArray(t *testing.T) {
	src := `message Msg {
	repeated string strings_list = 5 [
		// before
		(validate.rules).repeated.max_items = 20 // inline
		// after   
	];
}`
	p := newParserOn(src)
	msg := new(Message)
	p.next()
	if err := msg.parse(p); err != nil {
		t.Fatal(err)
	}
}
