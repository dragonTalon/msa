package command

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"msa/pkg/config"
	"msa/pkg/logic/provider"
	"msa/pkg/model"
	"strings"
)

func init() {
	RegisterCommand(&ListModel{})
	RegisterCommand(&SetModel{})
}

type ListModel struct {
}

func (l *ListModel) Name() string {
	return "models"
}

func (l *ListModel) Description() string {
	return "List all models"
}

func (l *ListModel) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	p := provider.GetProvider()
	if p == nil {
		return nil, fmt.Errorf("provider not found")
	}
	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	var modelNames []string
	for _, m := range models {
		modelNames = append(modelNames, m.Name)
	}
	return &model.CmdResult{
		Code: 0,
		Msg:  "success",
		Type: "list",
		Data: modelNames,
	}, nil
}

type SetModel struct {
}

func (s *SetModel) Name() string {
	return "model"
}

func (s *SetModel) Description() string {
	return "set model"
}

func (s *SetModel) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	p := provider.GetProvider()
	if p == nil {
		log.Errorf("provider not found")
		return nil, fmt.Errorf("provider not found")
	}
	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	checkModel := make(map[string]string)
	for _, m := range models {
		checkModel[m.Name] = m.Description
	}
	log.Infof("checkModel: %v args: %v", checkModel, args)
	if len(args) == 0 {
		return nil, fmt.Errorf("model not found")
	}
	m := strings.TrimSpace(args[0])
	if _, ok := checkModel[m]; !ok {
		return nil, fmt.Errorf("model %s not found", m)
	}
	config.GetLocalStoreConfig().Model = m
	err = config.SaveConfig()
	if err != nil {
		return nil, err
	}
	log.Infof("SetModel: %s", m)
	return &model.CmdResult{
		Code: 0,
		Msg:  "修改成功",
		Type: "boolean",
		Data: true,
	}, nil
}
