package transfer

import (
	"github.com/giskook/gotcp"
	"github.com/giskook/transfer/conn"
	"github.com/giskook/transfer/pkg"
	"github.com/giskook/transfer/protocol"
	"log"
)

type DownstreamPacket struct {
	buf []byte
}

func (this *DownstreamPacket) Serialize() []byte {
	return this.buf
}

type DownstreamProtocol struct {
}

func (this *DownstreamProtocol) ReadPacket(c *gotcp.Conn) (gotcp.Packet, error) {
	smconn := c.GetExtraData().(*conn.Conn)
	smconn.UpdateReadflag()

	buffer := smconn.GetBuffer()
	conn := c.GetRawConn()
	for {
		if smconn.ReadMore {
			data := make([]byte, 2048)
			readLengh, err := conn.Read(data)
			log.Printf("<Down IN>  %x\n", data[0:readLengh])
			if err != nil {
				return nil, err
			}

			if readLengh == 0 {
				return nil, gotcp.ErrConnClosing
			}
			buffer.Write(data[0:readLengh])

			//		return &DownstreamPacket{
			//			buf: data[0:readLengh],
			//		}, nil
		}
		cmdid, pkglen := protocol.CheckProtocol(buffer)
		log.Printf("protocol id %x length %d\n", cmdid, pkglen)

		pkgbyte := make([]byte, pkglen)
		buffer.Read(pkgbyte)
		log.Println("Read")
		switch cmdid {
		case protocol.PROTOCOL_DOWN_REQ_REGISTER:
			log.Println("PROTOCOL_DOWN_REQ_REGISTER")
			p := protocol.ParseDownRegister(pkgbyte)
			smconn.ReadMore = false

			return pkg.NewTransparentTransmissionPakcet(cmdid, p), nil
		case protocol.PROTOCOL_DOWN_REQ_CANCEL:
			log.Println("PROTOCOL_DOWN_REQ_CANCEL")
			p := protocol.ParseDownCancel(pkgbyte)
			smconn.ReadMore = false

			return pkg.NewTransparentTransmissionPakcet(cmdid, p), nil

		case protocol.PROTOCOL_DOWN_TRANSFER:
			p := protocol.ParseDownTransfer(pkgbyte)
			smconn.ReadMore = false

			return pkg.NewTransparentTransmissionPakcet(cmdid, p), nil

		case protocol.PROTOCOL_ILLEGAL:
			smconn.ReadMore = true
		case protocol.PROTOCOL_HALF_PACK:
			smconn.ReadMore = true
		}
	}
}
