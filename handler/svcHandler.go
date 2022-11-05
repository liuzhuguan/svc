package handler

import (
	"context"
	"git.imooc.com/coding-535/common"
	log "github.com/asim/go-micro/v3/logger"
	"github.com/liuzhuguan/svc/domain/model"
	"github.com/liuzhuguan/svc/domain/service"
	"github.com/liuzhuguan/svc/proto/svc"
)

type SvcHandler struct {
	//注意这里的类型是 ISvcDataService 接口类型
	SvcDataService service.ISvcDataService
}

// Call is a single request handler called via client.Call or the generated client code
func (e *SvcHandler) AddSvc(ctx context.Context, info *svc.SvcInfo, rsp *svc.Response) error {
	log.Info("Received *svc.AddSvc request: ", info)

	tarSvc := &model.Svc{}
	if err := common.SwapTo(info, tarSvc); err != nil {
		common.Error(err)
		return err
	}

	if err := e.SvcDataService.CreateSvcToK8s(info); err != nil {
		common.Error(err)
		return err
	} else {
		if _, err := e.SvcDataService.AddSvc(tarSvc); err != nil {
			common.Error(err)
			return err
		}
	}

	return nil
}

func (e *SvcHandler) DeleteSvc(ctx context.Context, req *svc.SvcId, rsp *svc.Response) error {
	log.Info("Received *svc.DeleteSvc request", req)

	serviceInfo, err := e.SvcDataService.FindSvcByID(req.Id)
	if err != nil {
		common.Error(err)
		return err
	}

	if err = e.SvcDataService.DeleteFromK8s(serviceInfo); err != nil {
		common.Error(err)
		return err
	}

	if err = e.SvcDataService.DeleteSvc(req.Id); err != nil {
		common.Error(err)
		return err
	}

	return nil
}

func (e *SvcHandler) UpdateSvc(ctx context.Context, req *svc.SvcInfo, rsp *svc.Response) error {
	log.Info("Received *svc.UpdateSvc request")
	//先更新k8s里面的数据
	if err := e.SvcDataService.UpdateSvcToK8s(req); err != nil {
		common.Error(err)
		return err
	}
	//查询数据库中的svc
	serviceInfo, err := e.SvcDataService.FindSvcByID(req.Id)
	if err != nil {
		common.Error(err)
		return err
	}
	//数据类型转换
	if err := common.SwapTo(req, serviceInfo); err != nil {
		common.Error(err)
		return err
	}
	//更新到数据中
	if err := e.SvcDataService.UpdateSvc(serviceInfo); err != nil {
		common.Error(err)
		return err
	}
	return nil
}

func (e *SvcHandler) FindSvcByID(ctx context.Context, req *svc.SvcId, rsp *svc.SvcInfo) error {
	log.Info("查找服务")

	svcModel, err := e.SvcDataService.FindSvcByID(req.Id)
	if err != nil {
		common.Error(err)
		return err
	}

	if err := common.SwapTo(svcModel, rsp); err != nil {
		common.Error(err)
		return err
	}
	return nil
}

func (e *SvcHandler) FindAllSvc(ctx context.Context, req *svc.FindAll, rsp *svc.AllSvc) error {
	log.Info("查询所有服务")

	allSvc, err := e.SvcDataService.FindAllSvc()
	if err != nil {
		common.Error(err)
		return err
	}
	//整理格式
	for _, v := range allSvc {
		svcInfo := &svc.SvcInfo{}
		if err := common.SwapTo(v, svcInfo); err != nil {
			common.Error(err)
			return err
		}
		rsp.SvcInfo = append(rsp.SvcInfo, svcInfo)
	}
	return nil
}
