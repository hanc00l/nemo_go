$(function () {
    let tasks_table = $('#tasks-table').DataTable(
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
    let vulnerability_table = $('#vulnerability-table').DataTable(
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
    let onlineuser_table = $('#onlineuser-table').DataTable(
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
                "url": "/onlineuser-list",
                "type": "post",
                "data": {start: 0, length: 100} //显示全部记录
            },
            columns: [
                {data: "index", title: "序号", width: "5%"},
                {data: "ip", title: "IP", width: "35%"},
                {data: 'login_time', title: '登录时间', width: '20%',},
                {data: 'update_time', title: '更新时间', width: '20%'},
                {data: 'update_number', title: '更新次数', width: '15%'},
            ]
        }
    );//end datatable
    get_count_data();
    get_user_workspace_list();
    $('#select_workspace').change(function () {
        change_user_workspace();
    });
    //定时刷新页面
    setInterval(function () {
        get_count_data();
        onlineuser_table.ajax.reload();
        vulnerability_table.ajax.reload();
        tasks_table.ajax.reload();
    }, 60 * 1000);
});

// 获取统计信息
function get_count_data() {
    //异步获取任务统计信息
    $.post("/dashboard", function (data) {
        $("#vulnerability_count").html(data['vulnerability_count']);
        $("#domain_count").html(data['domain_count']);
        $("#ip_count").html(data['ip_count']);
        $("#task_active").html("Task:&nbsp;" + data['task_active']);
        $('#worker_count').html("Worker:&nbsp;" + data['worker_count']);
    });
}

//  获取用户的工作空间
function get_user_workspace_list() {
    document.getElementById('li_workspace').style.visibility = 'visible';
    $.post("/workspace-user-list", function (data) {
        $("#select_workspace").empty();
        for (let i = 0; i < data.WorkspaceInfoList.length; i++) {
            $("#select_workspace").append("<option value='" + data.WorkspaceInfoList[i].workspaceId + "'>" + data.WorkspaceInfoList[i].workspaceName + "</option>")
        }
        $('#select_workspace').val(data.CurrentWorkspace);
    });
}

// 手工切换用户的工作空间
function change_user_workspace() {
    let newWorkspace = $("#select_workspace").val();
    $.post("/workspace-user-change", {"workspace": newWorkspace}, function (data) {
        if (data['status'] == 'success') {
            swal({
                    title: "切换工作空间成功！",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                },
                function () {
                    get_count_data();
                });
        } else {
            swal('Warning', "切换工作空间失败! " + data['msg'], 'error');
        }
    });
}
