package state

// type metrics struct {
// 	State *FileState `json:"state"`
// }

// func (m *StateManager) MonitorPath() string {
// 	return "/state_manager"
// }

// func (m *StateManager) Handler(w http.ResponseWriter, r *http.Request) {
// 	b, err := json.Marshal(
// 		&metrics{
// 			State: m.State,
// 		},
// 	)
// 	if err != nil {
// 		w.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
// 	} else {
// 		w.Write(b)
// 	}
// }
