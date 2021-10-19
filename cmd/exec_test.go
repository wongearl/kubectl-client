package cmd

import (
	"log"
	"testing"
)

func TestExec(t *testing.T) {
	// kubectl cp /tmp/local  api-test-77c6f9bf8c-nhhp5:/opt -n dev
	pod := Pod("plumber-68f986d7dc-4fxj5", "kube-system", "plumber")
	message, err := pod.KubectlExec([]string{"ls", "-l"})
	if err != nil {
		log.Print(err)
	}
	log.Println("message:", string(message))

	message, err = pod.KubectlExec([]string{"pwd"})
	if err != nil {
		log.Print(err)
	}
	log.Println("message:", string(message))
}
