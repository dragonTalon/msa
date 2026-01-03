package provider

import (
	"msa/pkg/logic/provider/siliconflow"
	"msa/pkg/model"
)

var _ LLMProvider = (*siliconflow.SiliconflowProvider)(nil)

func init() {
	RegisterProvider(model.Siliconflow, &siliconflow.SiliconflowProvider{})
}
