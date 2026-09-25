package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

var (
	ErrClientClosed = errors.New("plugin client is closed")
	ErrInvalidRPC   = errors.New("invalid rpc response")
)

type Client struct {
	reader  io.Reader
	writer  io.WriteCloser
	writeMu sync.Mutex

	nextID atomic.Uint64

	mu      sync.Mutex
	pending map[uint64]chan *Response
	closed  bool
	closeCh chan struct{}
}

func NewClient(r io.Reader, w io.WriteCloser) *Client {
	c := &Client{
		reader:  r,
		writer:  w,
		pending: make(map[uint64]chan *Response),
		closeCh: make(chan struct{}),
	}
	go c.readLoop()
	return c
}

func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClientClosed
	}
	c.mu.Unlock()

	id := c.nextID.Add(1)

	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return err
		}
		rawParams = b
	}

	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	ch := make(chan *Response, 1)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClientClosed
	}
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	c.writeMu.Lock()
	_, err = c.writer.Write(data)
	c.writeMu.Unlock()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closeCh:
		return ErrClientClosed
	case resp, ok := <-ch:
		if !ok || resp == nil {
			return ErrClientClosed
		}
		if resp.Error != nil {
			return resp.Error
		}
		if result != nil && len(resp.Result) > 0 {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return err
			}
		}
		return nil
	}
}

func (c *Client) Notify(method string, params any) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClientClosed
	}
	c.mu.Unlock()

	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return err
		}
		rawParams = b
	}

	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  rawParams,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err = c.writer.Write(data)
	return err
}

func (c *Client) readLoop() {
	reader := bufio.NewReader(c.reader)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var resp Response
			if unmarshalErr := json.Unmarshal(line, &resp); unmarshalErr == nil {
				if resp.ID != 0 {
					c.mu.Lock()
					ch, ok := c.pending[resp.ID]
					if ok {
						delete(c.pending, resp.ID)
					}
					c.mu.Unlock()

					if ok && ch != nil {
						ch <- &resp
					}
				}
			}
		}
		if err != nil {
			c.Close()
			return
		}
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.closeCh)

	for id, ch := range c.pending {
		delete(c.pending, id)
		close(ch)
	}
	c.mu.Unlock()

	var err error
	if c.writer != nil {
		err = c.writer.Close()
	}
	if closer, ok := c.reader.(io.Closer); ok {
		_ = closer.Close()
	}
	return err
}
