package alerter

// Attribute is an attribute such as tx.origin or msg.value
type Attribute struct {
	Parts []string
}

// SelectClause is a select clause of a statement
type SelectClause struct {
	Attributes []Attribute
}

// FromClause is a from clause of a statement
type FromClause struct {
	// NOTE: can currently only be an address
	Address string
}

// WhereClause is a where clause
type WhereClause struct {
}

// LimitClause is a limit clause
type LimitClause struct {
	Limit int64
}

// SinceClause is a since clause
type SinceClause struct {
	Since int64
}

// UntilClause is an until clause
type UntilClause struct {
	Until int64
}

// GroupByClause is a group by clause
type GroupByClause struct {
}

// Statement is a full EMQL statement
type Statement struct {
	Select  SelectClause
	From    FromClause
	Where   WhereClause
	Limit   LimitClause
	Since   SinceClause
	Until   UntilClause
	GroupBy GroupByClause
}
