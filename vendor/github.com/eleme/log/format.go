package log

import (
	"bytes"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

const (
	// LDate is the layout of date
	LDate = "2006-01-02"
	// LTime is the laytou of time
	LTime = "15:04:05"
	// LDatetime is the layout of datetime
	LDatetime = LDate + " " + LTime + ".999"
)

// Formatter represents a formatter of record.
type Formatter interface {
	ParseFormat(format string) error
	Format(record Record) []byte
	Colored() bool
	SetColored(colored bool)
}

// BaseFormatter describes the format of outputting log
type BaseFormatter struct {
	colored     bool
	tpl         *template.Template
	tagReplacer *strings.Replacer
	tags        []string
	funcMap     template.FuncMap
}

// NewBaseFormatter creates a BaseFormatter with given colored whether
// to color the output
func NewBaseFormatter(colored bool) *BaseFormatter {
	f := new(BaseFormatter)
	f.colored = colored
	f.AddTags(defaultTags...)
	f.initFuncMap()
	return f
}

// AddTags add replacer oldnew tags in formatter.
func (f *BaseFormatter) AddTags(tags ...string) {
	f.tags = append(f.tags, tags...)
	f.tagReplacer = strings.NewReplacer(f.tags...)
}

var rTagLong = regexp.MustCompile("{{ *([a-zA-Z_]+) *}}")
var tagShort = []byte("{{$1}}")
var defaultTags = []string{
	"{{}}", "{{.String}}",
	"{{level}}", "{{level .}}",
	"{{l}}", "{{l .}}",
	"{{date}}", "{{date .}}",
	"{{time}}", "{{time .}}",
	"{{datetime}}", "{{datetime .}}",
	"{{name}}", "{{name .}}",
	"{{pid}}", "{{pid .}}",
	"{{file_line}}", "{{file_line .}}",

	"{{app_id}}", "{{app_id .}}",
}

// ParseFormat parse the format of outputting log
//
// The default format is "{{ level }} {{ date }} {{ time }} {{ name }} {{}}"
//
// {{this is a placeholder}} which will be replaced by the actual content
//
// Available placeholders:
//	{{}}            The message provided by you e.g. l.Info(message)
//	{{ level }}     Log level in four UPPER-CASED letters e.g. INFO, WARN
//	{{ l }}         Log level in one UPPER-CASED letter e.g. I, W
//	{{ data }}      Date in format "2006-01-02"
//	{{ time }}      Time in format "15:04:05")
//	{{ datetime }}  Date and time in format "2006-01-02 15:04:05.999"
//	{{ name }}      Logger name
//	{{ pid }}       Current process ID
//	{{ file_line }} Filename and line number in format "file.go:12"
func (f *BaseFormatter) ParseFormat(format string) error {
	// {{ tag }} -> {{tag}}
	format = string(rTagLong.ReplaceAll([]byte(format), tagShort))

	format = f.tagReplacer.Replace(format)

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	t, err := template.New("tpl").Funcs(f.funcMap).Parse(format)
	if err != nil {
		return err
	}

	// TODO: validation

	f.tpl = t
	return nil
}

// Format formats a Record with set format
func (f *BaseFormatter) Format(record Record) []byte {
	var buf bytes.Buffer // TODO: use sync.Pool
	f.tpl.Execute(&buf, record)
	return buf.Bytes()
}

// Colored return is colored.
func (f *BaseFormatter) Colored() bool {
	return f.colored
}

// SetColored set the value of colored.
func (f *BaseFormatter) SetColored(colored bool) {
	f.colored = colored
}

// TODO: the 'if color then Paint' is ugly!!

func (f *BaseFormatter) _level(r Record) string {
	s := LevelName[r.Level()]
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _l(r Record) string {
	s := LevelName[r.Level()][0:1]
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _datetime(r Record) string {
	s := r.Now().Format(LDatetime)
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _date(r Record) string {
	s := r.Now().Format(LDate)
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _time(r Record) string {
	s := r.Now().Format(LTime)
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _name(r Record) string {
	s := r.Name()
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _pid(r Record) string {
	s := strconv.Itoa(os.Getpid())
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _appID(r Record) string {
	s := r.AppID()
	if s == "" {
		s = "-"
	}
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) _fileLine(r Record) string {
	s := r.Fileline()
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			s = s[i+1:]
			break
		}
	}
	if f.colored {
		s = f.Paint(r.Level(), s)
	}
	return s
}

func (f *BaseFormatter) initFuncMap() {
	f.funcMap = template.FuncMap{
		"date":      f._date,
		"time":      f._time,
		"datetime":  f._datetime,
		"l":         f._l,
		"level":     f._level,
		"name":      f._name,
		"pid":       f._pid,
		"file_line": f._fileLine,
		"app_id":    f._appID,
	}
}

// AddFuncMap used to add template.FuncMap in formater.
func (f *BaseFormatter) AddFuncMap(funcMap template.FuncMap) {
	for k, v := range funcMap {
		f.funcMap[k] = v
	}
}

// Paint used to Paint the log.
func (f *BaseFormatter) Paint(lv LevelType, s string) string {
	return painter(levelColor[lv], s)
}
