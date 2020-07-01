package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
	"time"

	"github.com/pkg/errors"
)

func (s *Server) StreamBlockByHeight(
	req *proto.BlockByHeightRequest,
	ss proto.HighwayService_StreamBlockByHeightServer,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHeight request spec %v, type = %s, heights = %v %v", req.Specific, req.GetType().String(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])
	if err := proto.CheckReqNCapBlocks(req); err != nil {
		logger.Error(err)
		return err
	}
	logger.Infof("Receive StreamBlockByHeight request spec %v, type = %s, heights = %v %v", req.Specific, req.GetType().String(), req.GetHeights()[0], req.GetHeights()[len(req.GetHeights())-1])
	g := NewBlkGetter(req, nil)
	blkRecv := g.Get(ctx, s)
	sent, err := SendWithTimeout(blkRecv, common.MaxTimeForSend, ss.Send)
	logger.Infof("[stream] Successfully sent %v block to client", sent)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) StreamBlockByHash(
	req *proto.BlockByHashRequest,
	ss proto.HighwayService_StreamBlockByHashServer,
) error {
	if req.GetCallDepth() > common.MaxCallDepth {
		return errors.Errorf("reach max calldepth %v ", req)
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.MaxTimePerRequest)
	defer cancel()
	ctx = WithRequestID(ctx, req)
	logger := Logger(ctx)
	logger.Infof("Receive StreamBlockByHash request, type = %s, hashes = %v %v", req.GetType().String(), req.GetHashes()[0], req.GetHashes()[len(req.GetHashes())-1])

	g := NewBlkGetter(nil, req)
	blkRecv := g.Get(ctx, s)
	sent, err := SendWithTimeout(blkRecv, common.MaxTimeForSend, ss.Send)
	logger.Infof("[stream] Successfully sent %v block to client", sent)
	if err != nil {
		return err
	}
	return nil
}

func SendWithTimeout(blkChan chan common.ExpectedBlk, timeout time.Duration, send func(*proto.BlockData) error) (uint, error) {
	errChan := make(chan error, 10)
	t := time.NewTimer(timeout)
	defer t.Stop()
	numOfSentBlk := uint(0)
	for blk := range blkChan {
		if len(blk.Data) == 0 {
			return numOfSentBlk, nil
		}
		go func() {
			errChan <- send(&proto.BlockData{Data: blk.Data})
		}()
		select {
		case <-t.C:
			return numOfSentBlk, errors.Errorf("[stream] Trying send to client but timeout")
		case err := <-errChan:
			if err != nil {
				return numOfSentBlk, err
			}
			numOfSentBlk++
		}
	}
	return numOfSentBlk, nil
}
