//выводит/скрывает процесс загрузки
//   progress_loading('#comments',show=1,text='загрузка',size=40)  //вывод
//   progress_loading('#comments',show=0)                          //скрыть
function progress_loading(id_selector,show,text,size){
	var obj = $(id_selector);
	if(!show){
		//obj.remove('#loading3000');
		obj.find('#loading3000').remove();
	}
	var p = obj.find('#loading3000');
	if(show && p.length==0){
		//alert('new progress')
		if(!size) size = 50;
		if(!text) text = '';
		obj.append('<div class="container" id=loading3000 style="text-align:center;">'+text+' <img src="/img/loading.gif" style="width:'+size+'px;"/></div>');
	}
}

//вывод сообщений об ошибках
var sys_errcnt_500 = 0;
function show_msg(id_selector,msg,opt){
	var defopt = {
			type: 'danger',
			max_msg_show: 1,
			is_sys_error: 0,         
			sys_error_max_cnt: 3,
			sys_error_add_text: "<br>сообщите об ошибке администратору сайта на mixamarciv@gmail.com"
		}
	if(!opt) opt = defopt;
	if(!opt.max_msg_show) opt.max_msg_show = defopt.max_msg_show;
	var obj = $(id_selector);
	if(obj.length==0) return alert('не найден объект по селектору "'+id_selector+'" в функции show_msg(id_selector="'+id_selector+'",msg="'+msg+'"...');
	var errsb = obj.find('.s_error3000');
	if(errsb.length >= opt.max_msg_show){
		var t = errsb.length;
		var i = 0;
		while(t>=opt.max_msg_show){
			t--;
			var el = errsb[i++];
			el.remove();
		}
	}
	
	if(opt.is_sys_error){
		sys_errcnt_500++;
		if(!opt.sys_error_max_cnt) opt.sys_error_max_cnt = defopt.sys_error_max_cnt;
		if(sys_errcnt_500>=opt.sys_error_max_cnt){
			if(!opt.sys_error_add_text) opt.sys_error_add_text = defopt.sys_error_add_text;
			msg += opt.sys_error_add_text;
		}
	}
	if(!opt.type || opt.type=='error') opt.type = 'danger';
	s = '<div class="alert alert-dismissible alert-'+opt.type+' s_error3000"><button type="button" class="close" data-dismiss="alert">&times;</button>'+msg+'</div>';
	obj.append(s);
}

function show_msg_arr(id_selector,amsg,opt){
	a = ''
	for(i=0;i<amsg.length;i++){
		a += amsg[i]+'<br>'
	}
	show_msg(id_selector,a,opt)
}
