(function($){
    $.fn.serializeJson=function(){
        var serializeObj={};
        var array=this.serializeArray();
        var str=this.serialize();
        $(array).each(function(){
            if(serializeObj[this.name]){
                if($.isArray(serializeObj[this.name])){
                    serializeObj[this.name].push(this.value);
                }else{
                    serializeObj[this.name]=[serializeObj[this.name],this.value];
                }
            }else{
                serializeObj[this.name]=this.value;
            }
        });
        return serializeObj;
    };
})(jQuery);
var hz = function(){

	/**
	 * 页面loading
	 */
	var pageLoader = function($mode) {
		var $loadingEl = jQuery('#lyear-loading');
		    $mode      = $mode || 'show';
		if ($mode === 'show') {
			if ($loadingEl.length) {
				$loadingEl.fadeIn(250);
			} else {
				jQuery('body').prepend('<div id="lyear-loading"><div class="spinner-border text-primary" role="status"><span class="sr-only">Loading...</span></div></div>');
			}
		} else if ($mode === 'hide') {
			if ($loadingEl.length) {
				$loadingEl.fadeOut(250);
			}
		}
		return false;
	};

    /**
     * 页面小提示
     * @param $msg 提示信息
     * @param $type 提示类型:'info', 'success', 'warning', 'danger'
     * @param $delay 毫秒数，例如：1000
     * @param $icon 图标，例如：'fa fa-user' 或 'glyphicon glyphicon-warning-sign'
     * @param $from 'top' 或 'bottom'
     * @param $align 'left', 'right', 'center'
     * @author CaiWeiMing <314013107@qq.com>
     */
    var tips = function ($msg, $type, $delay, $icon, $from, $align) {
        $type  = $type || 'info';
        $delay = $delay || 1000;
        $from  = $from || 'top';
        $align = $align || 'center';
        $enter = $type == 'danger' ? 'animated shake' : 'animated fadeInUp';

        jQuery.notify({
            icon: $icon,
            message: $msg
        },
        {
            element: 'body',
            type: $type,
            allow_dismiss: true,
            newest_on_top: true,
            showProgressbar: false,
            placement: {
                from: $from,
                align: $align
            },
            offset: 20,
            spacing: 10,
            z_index: 10800,
            delay: $delay,
            //timer: 1000,
            animate: {
                enter: $enter,
                exit: 'animated fadeOutDown'
            }
        });
    };

    /**
     * 校验input是否有值
     * @param $obj 提示信息
     * @param $isTip 提示类型:'info', 'success', 'warning', 'danger'
     * @return boolean
     * @author huzhaolun
     */
    var hasVale = function ($obj,$isTip){
        let val = $($obj).val().trim();
        if (val == null || val === "" || typeof val == "undefined"){
            if ($isTip){
                let tip = $($obj).attr("placeholder");
                tips(tip||'参数不能为空', 'danger', 5000, 'mdi mdi-emoticon-happy', 'top', 'center')
            }
            return false;
        }else {
            return true;
        }
    };

    var formDataLoad = function (domId, obj) {
        for (var property in obj) {
            if (obj.hasOwnProperty(property) == true) {
                if ($("#" + domId + " [name='" + property + "']").size() > 0) {
                    $("#" + domId + " [name='" + property + "']").each(function () {
                        var dom = this;
                        if ($(dom).attr("type") == "radio") {
                            $(dom).filter("[value='" + obj[property] + "']").attr("checked", true);
                        }
                        if ($(dom).attr("type") == "checkbox") {
                            obj[property] == true ? $(dom).attr("checked", "checked") : $(dom).attr("checked", "checked").removeAttr("checked");
                        }
                        if ($(dom).attr("type") == "text" || $(dom).prop("tagName") == "SELECT" || $(dom).attr("type") == "hidden" || $(dom).attr("type") == "textarea") {
                            $(dom).val(obj[property]);
                        }
                        if ($(dom).prop("tagName") == "TEXTAREA") {
                            $(dom).val(obj[property]);
                        }
                    });
                }
            }
        }
    }

    function renderTime(date) {
        var dateee = new Date(date).toJSON();
        return new Date(+new Date(dateee) + 8 * 3600 * 1000).toISOString().replace(/T/g, ' ').replace(/\.[\d]{3}Z/, '')
    }

	return {
        // 页面小提示
        notify  : function ($msg, $type, $delay, $icon, $from, $align) {
            tips($msg, $type, $delay, $icon, $from, $align);
        },
        // 页面加载动画
		loading : function ($mode) {
		    pageLoader($mode);
		},
        hasVale : function ($obj,$isTip){
            return hasVale($obj,$isTip);
        },
        formDataLoad : function(domId, obj) {
            return formDataLoad(domId, obj);
        },
        renderTime : function (date){
            return renderTime(date)
        }
    };
}();
