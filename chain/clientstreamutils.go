package chain

import (
	context "context"
	"highway/common"
	"highway/proto"
	"time"

	"github.com/pkg/errors"
)

func (c *Client) StreamBlkByHeight(
	ctx context.Context,
	req RequestBlockByHeight,
	blkChan chan common.ExpectedBlk,
) error {
	logger := Logger(ctx)
	defer close(blkChan)
	st := time.Now()
	logger.Infof("[dbgleakmem] StreamBlkByHeight Start")
	logger.Infof("[stream] Server call Client: Start stream request %v", req)
	sc, _, err := c.getClientWithBlock(ctx, int(req.GetFrom()), req.GetHeights()[len(req.GetHeights())-1])
	if sc == nil {
		logger.Warnf("[stream] Client is nil!")
		blkChan <- common.ExpectedBlk{
			Height: req.GetHeights()[0],
			Data:   []byte{},
		}
		logger.Infof("[dbgleakmem] StreamBlkByHeight End %v", time.Since(st))
		return nil
	}
	nreq, ok := req.(*proto.BlockByHeightRequest)
	if !ok {
		blkChan <- common.ExpectedBlk{
			Height: req.GetHeights()[0],
			Data:   []byte{},
		}
		logger.Infof("[dbgleakmem] StreamBlkByHeight End %v", time.Since(st))
		return errors.Errorf("Invalid Request %v", req)
	} else {
		nreq.CallDepth++
	}
	stream, err := sc.StreamBlockByHeight(ctx, nreq)
	if err != nil {
		logger.Infof("[stream] Server call Client return error %v", err)
		blkChan <- common.ExpectedBlk{
			Height: req.GetHeights()[0],
			Data:   []byte{},
		}
		logger.Infof("[dbgleakmem] StreamBlkByHeight End %v", time.Since(st))
		return err
	}
	defer stream.CloseSend()
	logger.Infof("[stream] Server call Client: OK, return stream %v", stream)
	heights := req.GetHeights()
	blkHeight := heights[0] - 1
	idx := 0
	blkData := new(proto.BlockData)
	for blkHeight < heights[len(heights)-1] {
		if req.GetSpecific() {
			blkHeight = heights[idx]
			idx++
		} else {
			blkHeight++
		}
		if err == nil {
			blkData, err = stream.Recv()
			if err == nil {
				logger.Infof("[stream] Received block %v", blkHeight)
				blkChan <- common.ExpectedBlk{
					Height: blkHeight,
					Data:   blkData.GetData(),
				}
				continue
			} else {
				logger.Infof("[stream] Received err %v %v", stream, err)
			}
		}
		blkChan <- common.ExpectedBlk{
			Height: blkHeight,
			Data:   []byte{},
		}
	}
	for {
		_, errStream := stream.Recv()
		if errStream != nil {
			logger.Infof("[stream] Stream received err %v", err)
			break
		}
	}
	logger.Infof("[dbgleakmem] StreamBlkByHeight End %v", time.Since(st))
	return nil
}
