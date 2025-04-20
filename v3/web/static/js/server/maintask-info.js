$(function () {
    load_maintask_data();
    $('#select_task_status').change(function () {
        load_executor_task_data($('#task_id').text());
    });
})

function fill_form_with_data(data) {
    $('#task_id').text(data.task_id);
    $("#task_name").text(data.name);
    $("#task_status").text(data.status);
    $("#task_profile_name").text(data.profile_name);
    $("#task_workspace_name").text(data.workspace);
    $("#task_org_name").text(data.org_name);
    $("#task_create_time").text(data.create_time);
    $("#task_start_time").text(data.start_time);
    $("#task_end_time").text(data.end_time);
    $("#task_run_time").text(data.runtime);
    $("#task_executors_num").text(data.progress);
    $("#task_progress").text(data.progress_rate);
    $("#task_result").text(data.result);
    $("#task_proxy").text(data.proxy);
    $("#task_cron").text(data.cron_task);
    $("#task_split").text(data.target_split);
    $("#task_description").text(data.description);
    $("#task_target").text(data.target);
    $("#task_exclude_target").text(data.exclude_target);
    $("#task_args").text(data.args);
    if (data.is_cron_task === true) {
        $('#task_executor_card').hide();
    }
}

function load_maintask_data() {
    const urlParams = new URLSearchParams(window.location.search);
    const id = urlParams.get('id'); // 假设URL参数名为id

    if (!isNotEmpty(id)) {
        alert("任务ID未提供");
        return;
    }
    $.ajax({
        type: 'POST',
        url: '/maintask-info', // Beego处理URL
        data: {id: id},
        dataType: 'json',
        success: function (response) {
            fill_form_with_data(response);
            load_executor_task_data(response.task_id);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function load_executor_task_data(maintaskId) {
    $.ajax({
        type: 'POST',
        url: '/maintask-executor-tasks', // Beego处理URL
        data: {maintaskId: maintaskId,taskStatus:$('#select_task_status').val()},
        dataType: 'json',
        success: function (response) {
            renderTaskList(response, maintaskId);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

// 渲染任务列表
function renderTaskList(tasks, maintaskId) {
    const tbody = $("#task_executor_tbody")
    tbody.empty(); // 清空现有内容

    tasks.forEach((task, index) => {
        const row = `
                    <tr>
                        <td>${index + 1}</td>
                        <td>${task.executor}</td>
                        <td>${task.status}</td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.target}</div></td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.result}</div></td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.start_time}</div></td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.end_time}</div></td>
                        <td>${task.runtime}</td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.worker}</div></td>
                        <td><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">${task.create_time}</div></td>
                        <td>
                            <a class="btn btn-sm btn-danger" href="javascript:delete_executor_ask('${task.id}','${maintaskId}')" role="button" title="删除">
                                <i class="fa fa-trash-o"></i>
                            </a>
                        </td>
                    </tr>
                `;
        tbody.append(row);
    });
}


function delete_executor_ask(id, maintaskId) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除该任务, 是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: '/maintask-executor-task-delete',
                data: {id: id},
                success: function (data) {
                    show_response_message("删除任务", data, function () {
                        load_executor_task_data(maintaskId);
                    })
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
}

function view_tree() {
    window.location.href = '/maintask-tree?maintaskId=' + $('#task_id').text();
}