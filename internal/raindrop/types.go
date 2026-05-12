package raindrop

type Ref struct {
	ID int64 `json:"$id"`
}

type Raindrop struct {
	ID         int64    `json:"_id"`
	Title      string   `json:"title"`
	Link       string   `json:"link"`
	Domain     string   `json:"domain"`
	Excerpt    string   `json:"excerpt"`
	Note       string   `json:"note"`
	Type       string   `json:"type"`
	Tags       []string `json:"tags"`
	Important  bool     `json:"important"`
	Collection Ref      `json:"collection"`
	Created    string   `json:"created"`
	LastUpdate string   `json:"lastUpdate"`
}

type Collection struct {
	ID     int64  `json:"_id"`
	Title  string `json:"title"`
	Count  int64  `json:"count"`
	Public bool   `json:"public"`
	Parent *Ref   `json:"parent,omitempty"`
	View   string `json:"view"`
	Color  string `json:"color"`
}

type Tag struct {
	ID    string `json:"_id"`
	Count int64  `json:"count"`
}

type Count struct {
	Count int64 `json:"count"`
}

type Filters struct {
	Result     bool  `json:"result"`
	Broken     Count `json:"broken"`
	Duplicates Count `json:"duplicates"`
	Important  Count `json:"important"`
	NoTag      Count `json:"notag"`
	Tags       []Tag `json:"tags"`
	Types      []Tag `json:"types"`
}

type UserStats struct {
	Result bool `json:"result"`
	Items  []struct {
		ID    int64 `json:"_id"`
		Count int64 `json:"count"`
	} `json:"items"`
	Meta struct {
		Pro        bool  `json:"pro"`
		ID         int64 `json:"_id"`
		Duplicates Count `json:"duplicates"`
		Broken     Count `json:"broken"`
	} `json:"meta"`
}

type raindropsResponse struct {
	Result bool       `json:"result"`
	Items  []Raindrop `json:"items"`
	Count  int64      `json:"count"`
}

type raindropResponse struct {
	Result bool     `json:"result"`
	Item   Raindrop `json:"item"`
}

type collectionsResponse struct {
	Result bool         `json:"result"`
	Items  []Collection `json:"items"`
}

type tagsResponse struct {
	Result bool  `json:"result"`
	Items  []Tag `json:"items"`
}
