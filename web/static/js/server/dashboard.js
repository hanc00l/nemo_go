function get_count_data() {
    //异步获取任务统计信息
    $.post("/dashboard", function (data) {
        $("#task_active").html(data['task_active']);
        $("#vulnerability_count").html(data['vulnerability_count']);
        $("#domain_count").html(data['domain_count']);
        $("#ip_count").html(data['ip_count']);
    });
}

$(function () {
    var table = $('#tasks-table').DataTable(
        {
            "rowID": 'uuid',
            "paging": false,
            "searching": false,
            "processing": true,
            "serverSide": true,
            "autowidth": true,
            "sort": false,
            "dom": '<t>',
            "ajax": {
                "url": "/task-list",
                "type": "post",
                "data": {start: 0, length: 5} //只显示最近5条记录
            },
            columns: [
                {data: 'task_name', title: '名称', width: '10%'},
                {data: 'state', title: '状态', width: '5%'},
                {
                    data: 'kwargs', title: '参数', width: '30%',
                    "render": function (data, type, row) {
                        var data = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                        return data;
                    }
                },
                {data: 'result', title: '结果', width: '15%'},
                {data: 'received', title: '接收时间', width: '10%'},
                {data: 'started', title: '启动时间', width: '10%'},
                {data: 'runtime', title: '执行时长', width: '8%'},
            ]
        }
    );//end datatable
    get_count_data();
    //定时刷新页面
    setInterval(function () {
        table.ajax.reload();
        get_count_data();
    }, 60 * 1000);
});


$(function () {
    var table = $('#vulnerability-table').DataTable(
        {
            "rowID": 'id',
            "paging": false,
            "searching": false,
            "processing": true,
            "serverSide": true,
            "autowidth": true,
            "sort": false,
            "dom": '<t>',
            "ajax": {
                "url": "/vulnerability-list",
                "type": "post",
                "data": {start: 0, length: 5} //只显示最近5条记录
            },
            columns: [
                {data: "target", title: "Target", width: "15%"},
                {data: "url", title: "URL", width: "20%"},
                {data: 'poc_file', title: 'Poc文件', width: '30%',},
                {data: 'source', title: '验证工具', width: '10%'},
                {data: 'update_datetime', title: '更新时间', width: '15%'},
            ]
        }
    );//end datatable
    get_count_data();
    //定时刷新页面
    setInterval(function () {
        table.ajax.reload();
        get_count_data();
    }, 60 * 1000);
});


$(function () {
    var table = $('#worker-table').DataTable(
        {
            "rowID": 'id',
            "paging": false,
            "searching": false,
            "processing": true,
            "serverSide": true,
            "autowidth": true,
            "sort": false,
            "dom": '<t>',
            "ajax": {
                "url": "/worker-list",
                "type": "post",
                "data": {start: 0, length: 5} //只显示最近5条记录
            },
            columns: [
                {data: "index", title: "序号", width: "5%"},
                {data: "worker_name", title: "Worker", width: "35%"},
                {data: 'create_time', title: '启动时间', width: '20%',},
                {data: 'update_time', title: '心跳时间', width: '20%'},
                {data: 'task_number', title: '已执行任务数', width: '15%'},
            ]
        }
    );//end datatable
    get_count_data();
    //定时刷新页面
    setInterval(function () {
        table.ajax.reload();
        get_count_data();
    }, 60 * 1000);
});

