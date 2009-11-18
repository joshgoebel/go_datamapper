package dm

import (
	"sqlite3";
	"fmt";
	"strings";
	"reflect";
	"container/vector";
	"os";
	"log";
	"time";
)

//type Connection sqlite3.Handle;

type Model struct {
	Name		string;
	TableName	string;
	Type		reflect.Type;
	Connection	*sqlite3.Handle;
	Table		TableSchema;
}

type TableSchema struct {
	Columns		[]Column;
}

type Column struct {
	Position int;
	Name string;
	DataType string;
}

type ResultSet struct {
	Results *vector.Vector;
}

type Opts map[string]interface{}

// func main()
// {
//
// }

const (
	SQLITE_INTEGER	= 1;
	SQLITE_TEXT	= 3;
	SQLITE_ROW	= 100;
)

var conn *sqlite3.Handle
var models map[string]Model
var logger *log.Logger

func AddModel(name string, table string, t reflect.Type) (m Model) {
	ts := build_table_schema(table);
	m = Model{name, table, t, conn, ts};
	models[name] = m;
	return;
}

func build_table_schema(table string) TableSchema {
	t := TableSchema{};
	tcs := []Column{Column{0, "id","int"}, 
		Column{1,"name","varchar(25)"}, Column{2,"abbv","varchar(2)"}};
	t.Columns = tcs;
	return t;
}

func Init(dbname string) {
	conn = new(sqlite3.Handle);
	models = make(map[string]Model);
	os.Mkdir("log", 0555);
	file, _ := os.Open("log/dm.log", os.O_WRONLY|os.O_CREAT|os.O_TRUNC, 0666);
	logger = log.New(file, nil, "", 0);
	r := conn.Open(dbname);
	if r != "" {
		println("ERROR")
	}

}

func (m *Model) single_result(sql string) interface{} {
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	o := reflect.MakeZero(m.Type);
	if err == 101 {	// not found
		f := o.(*reflect.StructValue).FieldByName("Null");
		f.(*reflect.BoolValue).Set(true);
		// return nil;
	} else {
		build_result(o, res)
	}

	return o.Interface();
}

func parse_options(opts ...) (options Opts) {
	v := reflect.NewValue(opts).(*reflect.StructValue);
	if v.NumField() > 0 {
		options = v.Field(0).Interface().(Opts)
	} else {
		options = Opts{}
	}
	return;
}

func camel_case(s string) string	{ return strings.ToUpper(s[0:1]) + s[1:len(s)] }

func (m *Model) build_results(st *sqlite3.Statement) (r *vector.Vector) {
	r = vector.New(0);
	for i := SQLITE_ROW; i == SQLITE_ROW; i = st.Step() {
		o := reflect.MakeZero(m.Type);
		x := build_result(o, st);
		r.Push(x.Interface());
	}
	return;
}

func build_result(o reflect.Value, st *sqlite3.Statement) reflect.Value {
	cc := st.ColumnCount();
	for i := 0; i < cc; i++ {
		cn := camel_case(st.ColumnName(i));
		ct := st.ColumnType(i);
		f := o.(*reflect.StructValue).FieldByName(cn);
		switch ct {
		case SQLITE_INTEGER:
			f.(*reflect.IntValue).Set(st.ColumnInt(i))
		case SQLITE_TEXT:
			f.(*reflect.StringValue).Set(st.ColumnText(i))
		}
	}
	return o;
}

func Execute(sql string) (s *sqlite3.Statement, err int) {
	errs := "";
	before := time.Nanoseconds();
	s, errs = conn.Prepare(sql);
	if errs != "" {
		println("SQL ERROR: " + conn.ErrMsg());
		logger.Logf("SQL ERROR  " +sql);
		die();
	}
	err = s.Step();
	after := time.Nanoseconds();
	elap := float64(after-before) / 1000 / 1000;
	logger.Logf("SQL (%.1fms)  "+sql, elap);
	return s, err;
}

func die()	{ os.Exit(1) }

func (m *Model) select_one_sql(id int, options Opts) string {
	return fmt.Sprintf("SELECT * FROM "+m.TableName+" WHERE id = %d", id)
}

func (m *Model) select_all_sql(options Opts) string {
	sql := "SELECT * FROM " + m.TableName;
	if len(options) > 0 {
		conds, ok := options["conditions"];
		if ok {
			sql = sql + fmt.Sprintf(" WHERE (%s)", conds.(string))
		}
		limit, ok := options["limit"];
		if ok && limit.(int) > 0 {
			sql = sql + fmt.Sprintf(" LIMIT %d", options["limit"].(int))
		}
	}
	return sql;
}

func (m *Model) select_count_sql(options Opts) string {
	return fmt.Sprintf("SELECT COUNT(id) FROM " + m.TableName)
}

// model stuff

func (m *Model) Count(opts ...) int {
	options := parse_options(opts);
	sql := m.select_count_sql(options);
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	return res.ColumnInt(0);
}

func (m *Model) All(opts ...) (r ResultSet) {
	options := parse_options(opts);
	r = ResultSet{};
	sql := m.select_all_sql(options);
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	r.Results = m.build_results(res);
	return;
}

func (m *Model) First(opts ...) interface{} {
	options := parse_options(opts);
	sql := m.select_all_sql(options) + " LIMIT 1";
	return m.single_result(sql);
}

func (m *Model) Find(id int, opts ...) interface{} {
	options := parse_options(opts);
	sql := m.select_one_sql(id, options);
	return m.single_result(sql);
}

func (m *Model) Last(opts ...) interface{} {
	options := parse_options(opts);
	sql := m.select_all_sql(options) + " ORDER BY id desc LIMIT 1";
	return m.single_result(sql);
}

func (m *Model) New() interface{} {
	o := reflect.MakeZero(m.Type);
	return o.Interface();
}

func (m *Model) Save(obj interface{}) bool {
	return m.Insert(obj);
}

func (m *Model) Insert(obj interface{}) bool {
	cs := make([]string, len(m.Table.Columns));
	vs := make([]string, len(m.Table.Columns));
	n := reflect.NewValue(obj);
	i := 0;
	for _, x := range m.Table.Columns {
		cs[i] = quote(x.Name);
		cn := camel_case(x.Name);
		f := n.(*reflect.StructValue).FieldByName(cn);
		if f!=nil {
			switch {
			case x.DataType=="int":
				v:= f.(*reflect.IntValue).Get();
				vs[i] = quote(fmt.Sprintf("%d",v));
			case strings.HasPrefix(x.DataType,"varchar"):
				vs[i] = quote(f.(*reflect.StringValue).Get());
			}
			i++;
		} 
		
	}
	columns := strings.Join(cs, ",");
	quoted_values := strings.Join(vs, ",");
	sql := "INSERT INTO " + m.TableName + " (" + columns + ") VALUES (" + quoted_values + ");";
	Execute(sql);
	return true;
}

func quote(s string) string {
	return "'" + s + "'";
}