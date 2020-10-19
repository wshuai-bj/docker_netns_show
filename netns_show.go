package main
 
import (
    "fmt"
    "os/exec"
    "strings"
    "io/ioutil"
)
 
var (
    NETNS_PATH string = "/var/run/netns"
    FILE_HEAD  string = "Docker_"
)

func Do_cmd(name string, arg ...string) string{
    cmd := exec.Command(name,arg...)
    output,err := cmd.Output()
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    //fmt.Println("Do_cmd output:",output,string(output))
    ret:=string(output)
    return ret
}


func get_docker_info() ([]string, []string) {
    CONTAINER_IDs   :=[]string{}
    IMAGEs          :=[]string{}

    ret := Do_cmd("docker", "ps")
    lines :=strings.Split(ret, "\n")
    // fmt.Println(len(lines))
    // for index,line := range(lines){
    //     fmt.Println(index,line)
    // }
    // 0 CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
    // 1 6d0088ad3e4b        centos              "bash"              3 minutes ago       Up 3 minutes                            happy_varahamihira
    // 2    

    if len(lines)>=3{
        lastLineIndex := len(lines)-1
        for index,line := range(lines){
            //fmt.Println(index,line)
            if index==0{
                continue
            }
            if index==lastLineIndex{
                continue
            }
            nodes := strings.Split(line, " ")
            CONTAINER_ID    := nodes[0]
            IMAGE           := ""
            for _index,node := range(nodes){
                if _index>0{
                    if len(node)>0{
                        IMAGE = node
                        break
                    }
                }
            }
            CONTAINER_IDs = append(CONTAINER_IDs,CONTAINER_ID)
            IMAGEs = append(IMAGEs,IMAGE)
        }
    }

    return CONTAINER_IDs,IMAGEs
}

func ListDir(path string) []string{
    files := []string{}
    ret,_:=ioutil.ReadDir(path)
    for _,f:= range(ret){
       name:= f.Name()
       files = append(files,name)
    }
    return files
}


func __delete_netns_dir(){
    files:=ListDir(NETNS_PATH)
    headLen := len(FILE_HEAD)
    for _,file:= range(files){
        if len(file) > headLen{
            if file[0:headLen]==FILE_HEAD{
                //fmt.Println(file,file[0:headLen])
                deletePath := NETNS_PATH+"/"+file
                fmt.Println("__delete_netns_dir() del ",deletePath)
                Do_cmd("rm", "-f",deletePath)
            }
        }
    }
}

func __getDockerPid(CONTAINER_ID string) string{
    //docker inspect --format '{{ .State.Pid }}' b004d3c3475e
    ret := Do_cmd("docker", "inspect","--format","'{{ .State.Pid }}'",CONTAINER_ID)
    //'28986'
    retLen := len(ret)
    //fmt.Println(ret,retLen)
    if retLen >2{
        ret = ret[1:retLen-2]
        //fmt.Println(ret,"ret")
    }else{
        ret = ""
    }

    return ret
}


//转义 / to ^
func __filter_save_ns_file_name(name string )string{
    ret :=strings.ReplaceAll(name, "/", "^")
    return ret
}

func save_docker_info(CONTAINER_IDs []string,IMAGEs []string){
    __delete_netns_dir()
    for index,CONTAINER_ID :=range(CONTAINER_IDs){
        pid     := __getDockerPid(CONTAINER_ID)
        IMAGE   := IMAGEs[index]
        from_ns_file_path   := "/proc/"+pid+"/ns/net"
        to_ns_file          := FILE_HEAD+IMAGE+"_"+CONTAINER_ID
        to_ns_file = __filter_save_ns_file_name(to_ns_file)
        to_ns_file_path     := NETNS_PATH+"/"+to_ns_file

        //ln -s /proc/19584/ns/net /var/run/netns/b004d3c3475e
        fmt.Println("save_docker_info() save ",to_ns_file)
        Do_cmd("ln", "-s",from_ns_file_path,to_ns_file_path)
    }
}


func main() {
    fmt.Println("reading.....docker")
    Do_cmd("mkdir", "-p" ,NETNS_PATH)

    CONTAINER_IDs,IMAGEs := get_docker_info()
    //fmt.Println(CONTAINER_IDs)
    //fmt.Println(IMAGEs)

    save_docker_info(CONTAINER_IDs,IMAGEs)
}
