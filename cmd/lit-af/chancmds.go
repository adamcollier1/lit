package main

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/fatih/color"
	"github.com/mit-dci/lit/litrpc"
	"github.com/mit-dci/lit/lnutil"
)

var fundCommand = &Command{
	Format: fmt.Sprintf("%s%s%s\n", lnutil.White("fund"),
		lnutil.ReqColor("peer", "coinType", "capacity", "initialSend"), lnutil.OptColor("data")),
	Description: fmt.Sprintf("%s\n%s\n%s\n%s\n",
		"Establish and fund a new lightning channel with the given peer.",
		"The capacity is the amount of satoshi we insert into the channel,",
		"and initialSend is the amount we initially hand over to the other party.",
		"data is an optional field that can contain 32 bytes of hex to send as part of the channel fund",
	),
	ShortDescription: "Establish and fund a new lightning channel with the given peer.\n",
}

var watchCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.White("watch"),
		lnutil.ReqColor("channel idx", "watchPeerIdx")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Send channel data to a watcher",
		"The watcher can defend your channel while you're offline."),
	ShortDescription: "Send channel watch data to watcher.\n",
}

var pushCommand = &Command{
	Format: fmt.Sprintf("%s%s%s%s\n", lnutil.White("push"), lnutil.ReqColor("channel idx", "amount"), lnutil.OptColor("times"), lnutil.OptColor("data")),
	Description: fmt.Sprintf("%s\n%s\n%s\n",
		"Push the given amount (in satoshis) to the other party on the given channel.",
		"Optionally, the push operation can be associated with a 32 byte value hex encoded.",
		"Optionally, the push operation can be repeated <times> number of times."),
	ShortDescription: "Push the given amount (in satoshis) to the other party on the given channel.\n",
}

var closeCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.White("close"), lnutil.ReqColor("channel idx")),
	Description: fmt.Sprintf("%s\n%s\n%s%s\n",
		"Cooperatively close the channel with the given index by asking",
		"the other party to finalize the channel pay-out.",
		"See also: ", lnutil.White("break")),
	ShortDescription: "Cooperatively close the channel with the given index by asking\n",
}

var breakCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.White("break"), lnutil.ReqColor("channel idx")),
	Description: fmt.Sprintf("%s\n%s\n%s%s\n",
		"Forcibly break the given channel. Note that you need to wait",
		"a set number of blocks before you can use the money.",
		"See also: ", lnutil.White("stop")),
	ShortDescription: "Forcibly break the given channel.\n",
}

var historyCommand = &Command{
	Format:           lnutil.White("history"),
	Description:      "Show all the metadata for justice txs",
	ShortDescription: "Show all the metadata for justice txs.\n",
}

func (lc *litAfClient) History(textArgs []string) error {
	if len(textArgs) > 0 && textArgs[0] == "-h" {
		fmt.Fprintf(color.Output, historyCommand.Format)
		fmt.Fprintf(color.Output, historyCommand.Description)
		return nil
	}

	args := new(litrpc.StateDumpArgs)
	reply := new(litrpc.StateDumpReply)

	err := lc.Call("LitRPC.StateDump", args, reply)
	if err != nil {
		return err
	}

	for _, tx := range reply.Txs {
		fmt.Fprintf(color.Output, "Pkh: %x, Idx: %d, Sig: %x, Txid: %x, Data: %x, Amt: %d\n", tx.Pkh, tx.Idx, tx.Sig, tx.Txid, tx.Data, tx.Amt)
	}

	return nil
}

func CheckHelpCommand(command *Command, textArgs []string, expectedLength int) error {
	if len(textArgs) > 0 && textArgs[0] == "-h" {
		fmt.Fprintf(color.Output, command.Format)
		fmt.Fprintf(color.Output, command.Description)
	}
	if len(textArgs) < expectedLength {
		// if number of args are less than expected, return
		return fmt.Errorf(command.Format)
	}
	return nil
}

func (lc *litAfClient) FundChannel(textArgs []string) error {
	err := CheckHelpCommand(fundCommand, textArgs, 4)
	if err != nil {
		return err
	}
	args := new(litrpc.FundArgs)
	reply := new(litrpc.StatusReply)

	peer, err := strconv.Atoi(textArgs[0])
	if err != nil {
		return err
	}
	coinType, err := strconv.Atoi(textArgs[1])
	if err != nil {
		return err
	}

	cCap, err := strconv.Atoi(textArgs[2])
	if err != nil {
		return err
	}
	iSend, err := strconv.Atoi(textArgs[3])
	if err != nil {
		return err
	}

	if len(textArgs) > 4 {
		data, err := hex.DecodeString(textArgs[4])
		if err != nil {
			// Wasn't valid hex, copy directly and truncate
			copy(args.Data[:], textArgs[3])
		} else {
			copy(args.Data[:], data[:])
		}
	}

	args.Peer = uint32(peer)
	args.CoinType = uint32(coinType)
	args.Capacity = int64(cCap)
	args.InitialSend = int64(iSend)

	err = lc.Call("LitRPC.FundChannel", args, reply)
	if err != nil {
		return err
	}

	fmt.Fprintf(color.Output, "%s\n", reply.Status)
	return nil
}

// Request close of a channel.  Need to pass in peer, channel index
func (lc *litAfClient) CloseChannel(textArgs []string) error {
	err := CheckHelpCommand(closeCommand, textArgs, 1)
	if err != nil {
		return err
	}

	args := new(litrpc.ChanArgs)
	reply := new(litrpc.StatusReply)

	cIdx, err := strconv.Atoi(textArgs[0])
	if err != nil {
		return err
	}

	args.ChanIdx = uint32(cIdx)

	err = lc.Call("LitRPC.CloseChannel", args, reply)
	if err != nil {
		return err
	}

	fmt.Fprintf(color.Output, "%s\n", reply.Status)
	return nil
}

// Almost exactly the same as CloseChannel.  Maybe make "break" a bool...?
func (lc *litAfClient) BreakChannel(textArgs []string) error {
	err := CheckHelpCommand(breakCommand, textArgs, 1)
	if err != nil {
		return err
	}

	args := new(litrpc.ChanArgs)
	reply := new(litrpc.StatusReply)

	cIdx, err := strconv.Atoi(textArgs[0])
	if err != nil {
		return err
	}

	args.ChanIdx = uint32(cIdx)

	err = lc.Call("LitRPC.BreakChannel", args, reply)
	if err != nil {
		return err
	}

	fmt.Fprintf(color.Output, "%s\n", reply.Status)
	return nil
}

// Push is the shell command which calls PushChannel
func (lc *litAfClient) Push(textArgs []string) error {
	err := CheckHelpCommand(pushCommand, textArgs, 2)
	if err != nil {
		return err
	}

	args := new(litrpc.PushArgs)
	reply := new(litrpc.PushReply)

	// this stuff is all the same as in cclose, should put into a function...
	cIdx, err := strconv.Atoi(textArgs[0])
	if err != nil {
		return err
	}
	amt, err := strconv.Atoi(textArgs[1])
	if err != nil {
		return err
	}

	times := int(1)
	if len(textArgs) > 2 {
		times, err = strconv.Atoi(textArgs[2])
		if err != nil {
			return err
		}
	}

	if len(textArgs) > 3 {
		data, err := hex.DecodeString(textArgs[3])
		if err != nil {
			// Wasn't valid hex, copy directly and truncate
			copy(args.Data[:], textArgs[3])
		} else {
			copy(args.Data[:], data[:])
		}
	}

	args.ChanIdx = uint32(cIdx)
	args.Amt = int64(amt)

	for times > 0 {
		err := lc.Call("LitRPC.Push", args, reply)
		if err != nil {
			return err
		}
		fmt.Fprintf(color.Output, "Pushed %s at state %s\n", lnutil.SatoshiColor(int64(amt)), lnutil.White(reply.StateIndex))
		times--
	}

	return nil
}

func (lc *litAfClient) Dump(textArgs []string) error {
	pReply := new(litrpc.DumpReply)
	pArgs := new(litrpc.NoArgs)

	err := lc.Call("LitRPC.DumpPrivs", pArgs, pReply)
	if err != nil {
		return err
	}
	fmt.Fprintf(color.Output, "Private keys for all channels and utxos:\n")

	// Display DumpPriv info
	for i, t := range pReply.Privs {
		fmt.Fprintf(color.Output, "%d %s h:%d amt:%s %s ",
			i, lnutil.OutPoint(t.OutPoint), t.Height,
			lnutil.SatoshiColor(t.Amt), t.CoinType)
		if t.Delay != 0 {
			fmt.Fprintf(color.Output, " delay: %d", t.Delay)
		}
		if !t.Witty {
			fmt.Fprintf(color.Output, " non-witness")
		}
		if len(t.PairKey) > 1 {
			fmt.Fprintf(
				color.Output, "\nPair Pubkey: %s", lnutil.Green(t.PairKey))
		}
		fmt.Fprintf(color.Output, "\n\tprivkey: %s", lnutil.Red(t.WIF))
		fmt.Fprintf(color.Output, "\n")
	}

	return nil
}

func (lc *litAfClient) Watch(textArgs []string) error {
	err := CheckHelpCommand(watchCommand, textArgs, 2)
	if err != nil {
		return err
	}

	args := new(litrpc.WatchArgs)
	reply := new(litrpc.WatchReply)

	cIdx, err := strconv.Atoi(textArgs[0])
	if err != nil {
		return err
	}

	peer, err := strconv.Atoi(textArgs[1])
	if err != nil {
		return err
	}

	args.ChanIdx = uint32(cIdx)
	args.SendToPeer = uint32(peer)

	err = lc.Call("LitRPC.Watch", args, reply)
	if err != nil {
		return err
	}

	fmt.Fprintf(color.Output, "Send channel %d data to peer %d\n",
		args.ChanIdx, args.SendToPeer)

	return nil
}
