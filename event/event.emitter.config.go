package event

const (
	MetaHash    = "hash"
	MetaTime    = "timestamp"
	MetaEvent   = "event"
	MetaVersion = "version"
	MetaDefault = "default"
)

type (
	EventConfig struct {
		Metadata map[string]map[string]interface{} `json:"metadata,omitempty" mapstructure:"metadata"`
		EventMap map[string]string                 `json:"event_map,omitempty" mapstructure:"event_map"`
		GroupMap map[string]string                 `json:"group_map,omitempty" mapstructure:"group_map"`
	}

	DriverConfig struct {
		Type   string      `json:"type" mapstructure:"type"`
		Config interface{} `json:"config" mapstructure:"config"`
	}
)

func NewEventConfig() *EventConfig {
	return &EventConfig{
		Metadata: make(map[string]map[string]interface{}),
		EventMap: make(map[string]string),
	}
}

func (c *EventConfig) getTopic(event string) string {
	if t, ok := c.EventMap[event]; ok {
		return t
	}
	return event
}

func (c *EventConfig) getGroup(group string) string {
	if g, ok := c.GroupMap[group]; ok {
		return g
	}
	return group
}

func (c *EventConfig) getMetadata(event string) map[string]interface{} {
	if m, ok := c.getMetadataCopy(event); ok {
		m[MetaEvent] = event
		return m
	}
	return c.getDefaultMetadata(event)
}

func (c *EventConfig) getDefaultMetadata(event string) map[string]interface{} {
	if m, ok := c.getMetadataCopy(MetaDefault); ok {
		m[MetaEvent] = event
		return m
	}

	return map[string]interface{}{
		MetaVersion: 1,
		MetaEvent:   event,
	}
}

//getMetadataCopy return copy of metadata map if available
func (c EventConfig) getMetadataCopy(name string) (map[string]interface{}, bool) {
	if m, ok := c.Metadata[name]; ok {
		copyMap := map[string]interface{}{}
		for k, v := range m {
			copyMap[k] = v
		}
		return copyMap, true
	}
	return nil, false
}
