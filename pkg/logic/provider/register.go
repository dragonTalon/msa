package provider

import (
	"msa/pkg/logic/provider/deepseek"
	"msa/pkg/logic/provider/siliconflow"
	"msa/pkg/model"
)

var _ LLMProvider = (*siliconflow.SiliconflowProvider)(nil)

var _ LLMProvider = (*deepseek.DeepseekProvider)(nil)

func init() {
	RegisterProvider(model.Siliconflow, &siliconflow.SiliconflowProvider{})
	RegisterProvider(model.Deepseek, &deepseek.DeepseekProvider{})
}
