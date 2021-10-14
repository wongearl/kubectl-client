package cmd

import (
	"log"
	"testing"
)

func TestCp(t *testing.T) {
	// kubectl cp /tmp/local  api-test-77c6f9bf8c-nhhp5:/opt -n dev
	pod := Pod("plumber-6f58cbc976-4nxmm", "kube-system", "plumber")
	//err := pod.Cp().ToPod("./test.txt", "/app")
	//if err != nil {
	//	log.Print(err)
	//}
	log.Println("=======================")
	//kubectl cp  api-test-775cf487ff-7zhnj:/opt/app.jar /tmp
	err2 := pod.Cp().FromPod("/app/plumber", ".")
	if err2 != nil {
		log.Print(err2)
	}
}
