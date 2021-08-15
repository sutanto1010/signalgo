class SignalGo {
    conn=null
    id=null
    connected=false
    options={
        //url:location.hostname+":"+location.port+"/ws",
        url:"26idvz710k.execute-api.ap-southeast-1.amazonaws.com/dev?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImMzbGM1YnFrcGprYzcycG5zMWVnIiwidWlkIjo5LCJlbWFpbCI6InN1dGFudG8uMTAxMEBnbWFpbC5jb20iLCJob3RlbHMiOm51bGwsInJvbGVzIjpudWxsLCJleHAiOiIyMDIxLTA3LTE4VDEwOjA2OjM5LjExNjY4NjU2OVoifQ.VjZgx3DrYiiwN144J1PW_YH8h2xep7RrauoW_4pr1n4",
        autoRetry:true,
        autoRetryInMs:500,
    }
    events={}
    OnClose=null
    OnError=null
    OnConnected=null
    log={
        i(msg){
            console.info(`${Date.now()}-signal-go: `+msg)
        },
        e(msg){
            console.error(`${Date.now()}-signal-go: `+msg)
        },
        d(msg){
            console.debug(`${Date.now()}-signal-go: `+msg)
        }
    }
    Connect(opts){
        let options=this.options
        this.options={
            ...options,
            ...opts
        }
        this.doConnect()
    }
    reConnect(){
        if(this.options.autoRetry){
            setTimeout(()=>{
                if(this.conn.readyState==WebSocket.CLOSED) {
                    this.doConnect()
                }
            }, this.options.autoRetryInMs)
        }
    }
    doConnect(){
        if(!this.connected){
            let idParam=""
            if(this.id){
                idParam=`?id=${this.id}`
            }
            if(this.conn!=null){
                delete(this.conn)
            }
            if(this.conn){
                this.conn.onclose=null
                this.conn.onerror=null
                this.conn.onopen=null
                this.conn.onmessage=null
            }
            let conn=new WebSocket(`wss://${this.options.url+idParam}`)
            conn.onclose=(ev) => {
                this.log.e("On Close: " + ev)
                this.connected = false
                if (this.OnClose) {
                    this.OnClose(ev)
                }
                conn=null
                this.reConnect()
            }

            conn.onmessage=(ev) => {
                console.log(ev)
                /*
                let reader = new FileReader()
                reader.onload=(e)=>{
                    let payload = JSON.parse(e.target.result)
                    let messageType=payload.t
                    let body = JSON.parse(payload.m)
                    if(messageType==1){
                        let callBack = this.events[payload.e]
                        if(callBack){
                            callBack(body)
                        }
                    }else if(messageType==0){
                        this.id=body
                    }
                    this.log.i(body)
                }
                reader.readAsText(ev.data)
                */
            }
            conn.onerror=(ev)=>{
                this.log.e("On Error: " + ev)
                this.connected = false
                if (this.OnError) {
                    this.OnError(ev)
                }
                conn=null
                this.reConnect()
            }
            conn.onopen= (ev) => {
                this.log.i("Connected: " + ev)
                this.connected = true
                if (this.OnConnected) {
                    this.OnConnected(ev)
                }
                Object.keys(this.events).forEach(key => {
                    this.registerEvent(key);
                })
            }
            this.conn=conn
        }
    }

    registerEvent(key) {
        let payload = {
            "t": 3,
            "e": key
        }
        this.doSend(payload, "Error on event registration!")
    }

    On(ev, callBack){
        this.events[ev]=callBack
        if(this.connected){
            this.registerEvent(ev)
        }
    }
    JoinGroup(group){
        let payload = {
            "t":2,
            "m":group
        }
        this.doSend(payload,"Can't join to any group if connection is not established yet!")
    }
    doSend(payload,err,force){
        if(!this.connected && !force){
            console.error(err)
            return
        }
        //let buff=this.str2ab(JSON.stringify(payload))
        this.conn.send(JSON.stringify(payload))
    }
    Send(ev,message){
        let payload = {
            "t":1,
            "m":message,
            "e":ev
        }
        this.doSend(payload,"Can't send message if connection is not established yet!")
    }
    ab2str(buf) {
        return String.fromCharCode.apply(null, new Uint16Array(buf));
    }
    str2ab(str) {
        let buf = new ArrayBuffer(str.length); // 2 bytes for each char
        let bufView = new Uint8Array(buf);
        for (var i=0, strLen=str.length; i < strLen; i++) {
            bufView[i] = str.charCodeAt(i);
        }
        return buf;
    }
}