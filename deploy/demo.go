package deploy

import (
	"boudle/client"
	"boudle/until"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Boudle struct {
	Uuid        int64    `json:"uuid,omitempty"`
	ChineseName string   `json:"chinese_name"`
	Describe    string   `json:"describe"`
	Moudle      []Moudle `json:"moudle"`
	CreateTime  int64    `json:"create_time,omitempty"`
}

type Moudle struct {
	EnglishName string        `json:"english_name"`
	ChineseName string        `json:"chinese_name"`
	Describe    string        `json:"describe"`
	Application []Application `json:"application"`
}

type Application struct {
	EnglishName string      `json:"english_name"`
	ChineseName string      `json:"chinese_name"`
	Describe    string      `json:"describe"`
	Environment Environment `json:"environment"`
}

type Environment struct {
	Language    string   `json:"language"`
	WorkDir     string   `json:"work_dir"`
	PackageList []string `json:"package_list"`
	Request     Request  `json:"request"`
}

type Request struct {
	Cpu    int32  `json:"cpu"`
	Memory string `json:"memory"`
}

func Create(c *gin.Context) {
	// 1. 获取前端发过的值绑定到结构体中
	var boudle Boudle
	if err := c.ShouldBind(&boudle); err == nil {
		fmt.Println(boudle)
	} else {
		fmt.Printf("err: ,%v\n", err)
	}
	boudle.Uuid = until.GenID()
	boudle.CreateTime = time.Now().Unix()
	fmt.Println(boudle.Uuid, boudle.CreateTime)
	// 资源保存到CRD中去

	// 创建镜像
	dir := strconv.FormatInt(boudle.Uuid, 10)
	image := createImage(dir, c)
	//判断镜像是否创建成功 成功就部署 不成功就退出
	if image == "" {
		return
	}
	// 部署资源
	deploy(image)
}
func createImage(dir string, c *gin.Context) string {
	// 1.创建文件夹
	dstName := "/root/remote-project-file/" + dir
	srcName := "/root/remote-project-file/Dockerfile"
	err := os.MkdirAll(dstName, os.ModePerm)
	if err != nil {
		fmt.Println("create dir failed, err:", err)
		return ""
	}
	// 2.接收文件
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return " "
	}
	files := form.File["files"]
	var projectDir string
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		path := strings.Split(file.Filename, filename)[0]
		projectDir = strings.Split(path, "/")[0]
		err := os.MkdirAll(dstName+"/"+path, os.ModePerm)
		if err != nil {
			fmt.Println("create dir failed, err:", err)
			return ""
		}
		if err = c.SaveUploadedFile(file, dstName+"/"+file.Filename); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return " "
		}
	}
	// 3.复制DockerFile到新的文件夹内
	dst := dstName + "/" + projectDir + "/"
	dstFile := dst + "Dockerfile"
	_, err = until.CopyFile(dstFile, srcName)
	if err != nil {
		fmt.Println("copy file failed, err:", err)
		return " "
	}
	// 4. 执行docker build -t 命令
	//更改工作目录
	err = os.Chdir(dst)
	if err != nil {
		fmt.Println("cd failed", err)
		return ""
	}
	image := dir + ":v1"
	out, err := until.Cmd("docker", "build", "-t", image, ".")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println("the result :", out)
	return image
}

func deploy(images string) {
	deployWordpressDeployment(images)
	deployMysqlDeployment()
	deployWordpressSvc()
	deployMysqlSvc()
}

func deployWordpressDeployment(images string) {
	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wordpress-deployment",
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "wordpress"},
			},
			Replicas: int32Ptr(1),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "wordpress",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "wordpress",
							Image: images,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(80),
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := client.Clientset.AppsV1().Deployments(corev1.NamespaceDefault).Create(context.Background(), deploy, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("create deploy failed ,err:%v\n", err)
		return
	}
}

func deployWordpressSvc() {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wordpress",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:     int32(8080),
					NodePort: int32(30001),
					TargetPort: intstr.IntOrString{
						IntVal: int32(80),
					},
				},
			},
			Selector: map[string]string{
				"app": "wordpress",
			},
		},
	}
	svc, err := client.Clientset.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("create wordpress failed, err: %v\n", err)
		return
	}
	fmt.Println(svc)
}

func deployMysqlDeployment() {
	var hostPathDir = corev1.HostPathDirectoryOrCreate
	var hostPathFile = corev1.HostPathFileOrCreate
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mysql",
		},
		Spec: v1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
				"app": "mysql",
			}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "mysql",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: "mysql:5.7",
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"/var/lib/mysql/init.sh",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mysql",
									MountPath: "/var/lib/mysql/",
								},
								{
									Name:      "init",
									MountPath: "/var/lib/mysql/init.sh",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: "secret",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(3306),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "mysql",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mysql",
									Type: &hostPathDir,
								},
							},
						},
						{
							Name: "init",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "//root/remote-project-file/example/php-mysql/init.sh",
									Type: &hostPathFile,
								},
							},
						},
					},
				},
			},
		},
	}
	deploy, err := client.Clientset.AppsV1().Deployments(corev1.NamespaceDefault).Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("create mysql deployment failed ,err:%v\n", err)
		return
	}
	fmt.Println(deploy)
}

func deployMysqlSvc() {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mysql",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       int32(3306),
					TargetPort: intstr.IntOrString{IntVal: int32(3306)},
				},
			},
			Selector: map[string]string{
				"app": "mysql",
			},
		},
	}
	svc, err := client.Clientset.CoreV1().Services(corev1.NamespaceDefault).Create(context.Background(), service, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("create mysql service failed ,err:%v/n", err)
		return
	}
	fmt.Println("mysql service is", svc)
}

func int32Ptr(i int32) *int32 { return &i }
