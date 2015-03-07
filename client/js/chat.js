const MESSAGE_TYPE_USER = 4;//用户消息
const MESSAGE_TYPE_GROUP = 3;//群组消息

function Chat (option) {
	// 配置参数
	this.host = option.host;
	this.app_key = option.app_key; //app key
	this.app_secret = option.app_secret; //app secret
	this.user = option.user; //用户标志
	this.password = option.password; //用户密码

	//websocket
	this.ws = new WebSocket(this.host);

	console.log('begin connect....')
	this.ws.onopen = function(){
		console.log('connect success....');

		if(option.connectSuccCallback)
		{
			option.connectSuccCallback();
		}
	}

	this.ws.onmessage = function(event){
		console.log(event.data)
	}

	this.ws.onclose = function(){
		console.log('connection close...');

		if(option.connectCloseCallback)
		{
			option.connectCloseCallback();
		}
	}

	this.ws.onerror = function(){
		console.log('socket error');
	}

	if(this.ws.readyState === undefined || this.ws.readyState > 1)
	{
		console.log('connect failed')
		if(option.connectFailCallback)
		{
			option.connectFailCallback();
		}
	}
}

//发送聊天消息
Chat.prototype.sendToUser = function(content, to) {

	var date = new Date();

	msg = {}
	msg.Cmd = 'MSG';
	msg.Params = {};
	msg.Params.UniqueId = date.getTime();
	msg.Params.Content = content;
	msg.Params.To = to;
	msg.Params.Type = MESSAGE_TYPE_USER;

	this.ws.send(msg.toJSONString())
};

//发送聊天消息
Chat.prototype.sendToGroup = function(content, to) {

	var date = new Date();

	msg = {}
	msg.Cmd = 'MSG';
	msg.Params = {};
	msg.Params.UniqueId = date.getTime();
	msg.Params.Content = content;
	msg.Params.To = to;
	msg.Params.Type = MESSAGE_TYPE_GROUP;

	this.ws.send(msg.toJSONString())
};