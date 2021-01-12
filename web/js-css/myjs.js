    //格式化年月日时分秒 
formatDate=function (){
        var date = new Date();
        var year = date.getFullYear();
        var month = date.getMonth()+1;
        month = month<10?'0'+month:month;
        var day = date.getDate();
        day = day<10?'0'+day:day;
        var hours = date.getHours();
        hours = hours<10?'0'+hours:hours;
        var minutes = date.getMinutes();
        minutes = minutes<10?'0'+minutes:minutes;
        var seconds = date.getSeconds();
        seconds = seconds<10?'0'+seconds:seconds;
        // 2019-07-23 09:40:30
        return year+'-'+month+'-'+day+' '+hours+':'+minutes+':'+seconds;
    }