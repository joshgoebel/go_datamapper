package dm

import (
	"sqlite3";
	"fmt";
	"strings";
	"reflect";
	"container/vector";
	"os";
)

//type Connection sqlite3.Handle;

type Model struct {
	Name		string;
	TableName	string;
	Type		reflect.Type;
	Connection	* sqlite3.Handle;
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

func AddModel(name string, table string, t reflect.Type) (m Model){
	m = Model{name, table, t, conn};
	models[name] = m;
	return
}

func Init(dbname string) {
	conn = new(sqlite3.Handle);
	models = make(map[string] Model);
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
	if err == 101 {// not found
		f := o.(*reflect.StructValue).FieldByName("Null");
		f.(*reflect.BoolValue).Set(true);
		// return nil;
		} else {
		build_result(o, res);
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
	// println(sql);
	s, errs = conn.Prepare(sql);
	if errs != "" {
		println("SQL ERROR: " + conn.ErrMsg());
		die();
	}
	err = s.Step();
	return s, err;
}

func die()	{ os.Exit(1) }

func (m *Model) select_one_sql(id int, options Opts) string {
	return fmt.Sprintf("select * from "+m.TableName+" where id = %d", id)
}

func (m *Model) select_all_sql(options Opts) string	{ 
	sql := "select * from " + m.TableName;
	if len(options)>0 {
		conds, ok := options["conditions"];
		if ok {
			sql = sql + fmt.Sprintf(" where (%s)", conds.(string))
		}	
		limit, ok := options["limit"];
		if ok && limit.(int) > 0 {
			sql = sql + fmt.Sprintf(" limit %d", options["limit"].(int))
		}	
	}
	return sql 
}

func (m *Model) select_count_sql(options Opts) string {
	return fmt.Sprintf("select count(id) from " + m.TableName)
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
	sql := m.select_all_sql(options) + " limit 1";
	return m.single_result(sql);	
}

func (m *Model) Find(id int, opts ...) interface{} {
	options := parse_options(opts);
	sql := m.select_one_sql(id, options);
	return m.single_result(sql);
}

func (m *Model) Last(opts ...) interface{} {
	options := parse_options(opts);
	sql := m.select_all_sql(options) + " order by id desc limit 1";
	return m.single_result(sql);
}
