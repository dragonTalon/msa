/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "msa/cmd"

func main() {
	cmd.Execute()
	//ask, err := agent.Ask(context.Background(), "美团股票分析", nil)
	//if err != nil {
	//	log.Errorf("ask failed: %v", err)
	//	return
	//}
	//stringbuilder := strings.Builder{}
	//for {
	//	message, err := ask.Recv()
	//	if err == io.EOF {
	//		fmt.Printf("done: %v\n", message)
	//		fmt.Printf("123" + stringbuilder.String())
	//		return
	//	}
	//	if message.ToolCalls != nil {
	//		for _, call := range message.ToolCalls {
	//			stringbuilder.WriteString(call.Function.Arguments)
	//			fmt.Printf("call: %v\n", utils.ToJSONString(call))
	//		}
	//	}
	//
	//	fmt.Printf("recv: %v\n", utils.ToJSONString(message))
	//
	//}

}
