package main

import "fmt"
import "encoding/gob"
import "bytes"
import "labix.org/v1/vclock"

//This is the Global Variable Struct that holds precious info
type GoVec struct {
	//processname 
	processname	string
	//vector clock in bytes
	currentVC [32]byte
	//should I print the log on screen every time I store it?
	printonscreen bool
	//should I assume that only I am logging (or are other remote hosts
	//doing so as well
	locallogging bool
	//current time for clock
	currenttime uint64
}

type Data struct {
    id int32
    name [32]byte
	vcinbytes [32]byte
	programdata [32]byte
}


func (d *Data) GobEncode() ([]byte, error) {
    w := new(bytes.Buffer)
    encoder := gob.NewEncoder(w)
    err := encoder.Encode(d.id)
    if err!=nil {
        return nil, err
    }
    err = encoder.Encode(d.name)
    if err!=nil {
        return nil, err
    }
	err = encoder.Encode(d.vcinbytes)
	if err!=nil {
		return nil,err
	}
	err = encoder.Encode(d.programdata)
	if err!=nil {
		return nil,err
	}
    return w.Bytes(), nil
}

func (d *Data) GobDecode(buf []byte) error {
    r := bytes.NewBuffer(buf)
    decoder := gob.NewDecoder(r)
    err := decoder.Decode(&d.id)
    if err!=nil {
        return err
    }
	decoder.Decode(&d.name)
	decoder.Decode(&d.vcinbytes)
    return decoder.Decode(&d.programdata)
}

func (d *Data) PrintDataBytes() {
	fmt.Println(d.id)
	fmt.Printf("%x \n",d.name)
	fmt.Printf("%X \n",d.vcinbytes)
}






func (gv *GoVec) PrepareSend(buf []byte) ([]byte){
/*
	This function is meant to be called before sending a packet. Usually,
	it should Update the Vector Clock for its own process, package with the clock
	using gob support and return the new byte array that should be sent onwards
	using the Send Command
*/
	//Converting Vector Clock from Bytes and Updating the gv clock
	vc:=FromBytes(gv.currentVC)
	gv.currenttime++
	vc.Update(gv.processname,gv.currenttime)
	gv.currentVC=vc.Bytes()
	//WILL HAVE TO CHECK THIS OUT!!!!!!! 
	
	//lets log the event
	//print
	
	//if only local logging the next is unnecceary since we can simply return buf as is 
	if gv.locallogging == true {
	//if true, then we add relevant info and encode it
		// Create New Data Structure and add information: data to be transfered
		d := Data{id:15} 
		copy(d.name[:], []byte(gv.processname))
		copy(d.vcinbytes[:],gv.currentVC)
		copy(d.programdata[:], buf)
		
		//create a buffer to hold data and Encode it
		buffer := new(bytes.Buffer)
		enc := gob.NewEncoder(buffer)
        err := enc.Encode(d)
		if err!=nil {
			panic(err)
		}
		//return buffer bytes which are encoded as gob. Now these bytes can be sent off and 
		// received on the other end!
		return buffer.Bytes()
	}
	// if we are only performing local logging, we have updated vector clock and logged it buffer can
	// be returned as is
	return buf	
}

func (gv *GoVec) UnpackRecieve(buf []byte) ([] byte){ 
/*
	This function is meant to be called immediatly after receiving a packet. It unpacks the data 
	by the program, the vector clock. It updates vector clock and logs it. and returns the user data
*/
	//if only our program is logging there is nothing attached in the byte buffer
	if gv.locallogging ==true {
		//simple adding to current time
		vc:=FromBytes(gv.currentVC)
	    gv.currenttime++
	    vc.Update(gv.processname,gv.currenttime)
		gv.currentVC=vc.Bytes()
		return buf
	}
	
	//if we are receiving a packet with a time stamp then:
    //Create a new buffer holder and decode into E, a new Data Struct
	
	//buffer = bytes.NewBuffer(buf)
	e := new(Data)
	//Decode Relevant Data... if fail it means that this doesnt not hold vector clock (probably)
	dec := gob.NewDecoder(buf)
    err := dec.Decode(e)
	if err!= nil {
		fmt.Println("You said that I would be receiving a vector clock but I didnt! or decoding failed :P")
		panic(err)
	}
	//In this case you increment your old clock
	vc:=FromBytes(gv.currentVC)
	gv.currenttime++
	vc.Update(gv.processname,gv.currenttime)
	//merge it with the new clock
	vc.Merge(FromBytes(e.vcinbytes))
	
	//Log it
	
    //  Out put the recieved Data
	return e.programdata
}

func New() *GoVec {
	return &GoVec{}
}

func (gv *GoVec) Initilize(n string, p bool , l bool){
/*This is the Start Up Function That should be called right at the start of 
a program

Assumption that Code Creates Logger Struct using :
Logger := GoVec.New()
Logger.Initialize(nameofprocess, printlogline, locallogging)
*/
	gv.processname = n
	gv.printonscreen = p
	gv.locallogging = l
	gv.currenttime = 0
	
	//we create a new Vector Clock with processname and 0 as the intial time
	vc1 := vlclock.New()
	vc1.Update(n, gv.currenttime)
	
	//Vector Clock Stored in bytes
	gv.currentvc=vc1.Bytes()

}


func main() {	
    vc1 := vclock.New()
	vc1.Update("A", 1)
    d := Data{id: 7}
    copy(d.name[:], []byte("tree"))
	copy(d.vcinbytes[:],vc1.Bytes())
	
	d.PrintDataBytes()
    buffer := new(bytes.Buffer)
    // writing
    enc := gob.NewEncoder(buffer)
    err := enc.Encode(d)
    if err != nil {
        fmt.Println("error")
    }
    // reading
    buffer = bytes.NewBuffer(buffer.Bytes())
    e := new(Data)
    dec := gob.NewDecoder(buffer)
    err = dec.Decode(e)
    //fmt.Println(e, err)
	e.PrintDataBytes()
}