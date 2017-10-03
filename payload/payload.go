package payload

type Payload struct {
	Size    int64     `json:"total_size,required"`
	Count   int64     `json:"record_count,required"`
	Records []*Record `json:"-"`
}

func NewPayload() *Payload {
	return &Payload{
		Size:    0,
		Count:   0,
		Records: make([]*Record, 0),
	}
}

func (p *Payload) AddRecord(r *Record) {
	p.Records = append(p.Records, r)
	p.Count++
	p.Size += r.Size
}

func (p *Payload) LastRecord() *Record {
	if p.Count == 0 {
		r := NewRecord()
		p.Records = append(p.Records, r)
		p.Count++
		return r
	}

	return p.Records[p.Count-1]
}
