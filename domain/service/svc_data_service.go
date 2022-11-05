package service

import (
	"context"
	"errors"
	"git.imooc.com/coding-535/common"
	"github.com/liuzhuguan/svc/domain/model"
	"github.com/liuzhuguan/svc/domain/repository"
	"github.com/liuzhuguan/svc/proto/svc"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

//这里是接口类型
type ISvcDataService interface {
	AddSvc(*model.Svc) (int64, error)
	DeleteSvc(int64) error
	UpdateSvc(*model.Svc) error
	FindSvcByID(int64) (*model.Svc, error)
	FindAllSvc() ([]model.Svc, error)
	CreateSvcToK8s(*svc.SvcInfo) error
	UpdateSvcToK8s(*svc.SvcInfo) error
	DeleteFromK8s(*model.Svc) error
}

//创建
//注意：返回值 ISvcDataService 接口类型
func NewSvcDataService(svcRepository repository.ISvcRepository, clientSet *kubernetes.Clientset) ISvcDataService {
	return &SvcDataService{SvcRepository: svcRepository, K8sClientSet: clientSet}
}

type SvcDataService struct {
	//注意：这里是 ISvcRepository 类型
	SvcRepository repository.ISvcRepository
	K8sClientSet  *kubernetes.Clientset
}

func (u *SvcDataService) CreateSvcToK8s(info *svc.SvcInfo) error {
	service := u.setService(info)

	if _, err := u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Get(context.TODO(), info.SvcName, v12.GetOptions{}); err != nil {
		//查找不到,就创建
		if _, err = u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Create(context.TODO(), service, v12.CreateOptions{}); err != nil {
			common.Error(err)
			return err
		}
		return nil
	} else {
		common.Error("Service " + info.SvcName + "已经存在")
		return errors.New("Service " + info.SvcName + "已经存在")
	}
}

func (u *SvcDataService) UpdateSvcToK8s(info *svc.SvcInfo) error {
	service := u.setService(info)

	if _, err := u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Get(context.TODO(), info.SvcName, v12.GetOptions{}); err != nil {
		common.Error("Service " + info.SvcName + "not exist")
		return errors.New("Service " + info.SvcName + "not exist")
	} else {
		if _, err = u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Update(context.TODO(), service, v12.UpdateOptions{}); err != nil {
			common.Error(err)
			return err
		}
		return nil
	}
}

func (u *SvcDataService) DeleteFromK8s(info *model.Svc) error {
	if _, err := u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Get(context.TODO(), info.SvcName, v12.GetOptions{}); err != nil {
		common.Error("Service " + info.SvcName + "not exist")
		return errors.New("Service " + info.SvcName + "not exist")
	} else {
		if err = u.K8sClientSet.CoreV1().Services(info.SvcNamespace).Delete(context.TODO(), info.SvcName, v12.DeleteOptions{}); err != nil {
			common.Error(err)
			return err
		}
		return nil
	}
}

//插入
func (u *SvcDataService) AddSvc(svc *model.Svc) (int64, error) {
	return u.SvcRepository.CreateSvc(svc)
}

//根据svcnfo 设置Iservice 信息
func (u *SvcDataService) setService(svcInfo *svc.SvcInfo) *v1.Service {
	service := &v1.Service{}
	//设置服务类型
	service.TypeMeta = v12.TypeMeta{
		Kind:       "v1",
		APIVersion: "Service",
	}
	//设置服务基础信息
	service.ObjectMeta = v12.ObjectMeta{
		Name:      svcInfo.SvcName,
		Namespace: svcInfo.SvcNamespace,
		Labels: map[string]string{
			"app-name": svcInfo.SvcPodName,
			"author":   "Caplost",
		},
		Annotations: map[string]string{
			"k8s/generated-by-lzg": "由lzg老师代码创建",
		},
	}
	//设置服务的spec信息，课程中采用ClusterIP模式
	service.Spec = v1.ServiceSpec{
		Ports: u.getSvcPort(svcInfo),
		Selector: map[string]string{
			"app-name": svcInfo.SvcPodName,
		},
		Type: "ClusterIP",
	}
	return service
}

func (u *SvcDataService) getSvcPort(svcInfo *svc.SvcInfo) (servicePort []v1.ServicePort) {
	for _, v := range svcInfo.SvcPort {
		servicePort = append(servicePort, v1.ServicePort{
			Name:       "port-" + strconv.FormatInt(int64(v.SvcPort), 10),
			Protocol:   v1.Protocol(v.SvcPortProtocol),
			Port:       v.SvcPort,
			TargetPort: intstr.FromInt(int(v.SvcTargetPort)),
		})
	}
	return
}

//删除
func (u *SvcDataService) DeleteSvc(svcID int64) error {
	return u.SvcRepository.DeleteSvcByID(svcID)
}

//更新
func (u *SvcDataService) UpdateSvc(svc *model.Svc) error {
	return u.SvcRepository.UpdateSvc(svc)
}

//查找
func (u *SvcDataService) FindSvcByID(svcID int64) (*model.Svc, error) {
	return u.SvcRepository.FindSvcByID(svcID)
}

//查找
func (u *SvcDataService) FindAllSvc() ([]model.Svc, error) {
	return u.SvcRepository.FindAll()
}
