$(function () {
    init_notify_multiselect();
    const id = $('#hidden_id').val();
    if (id !== "") {
        load_workspace_data(id);
    } else {
        load_notify_data(null);
    }
    $('#workspaceForm').on('submit', function (e) {
        e.preventDefault();
        save_workspace_data();
    });
});

function load_workspace_data(id) {
    $.ajax({
        type: 'post',
        url: '/workspace-get',
        data: {id: id,},
        dataType: 'json',
        success: function (data) {
            $('#workspace_name').val(data.workspace_name);
            $('#sort_number').val(data.sort_number);
            $('#status').val(data.status);
            $('#workspace_description').val(data.workspace_description);
            const selectedData = data.notify.split(',');
            load_notify_data(selectedData);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}

function save_workspace_data() {
    const workspace_name = $("#workspace_name").val();
    const status = $("#status").val();
    const sort_number = $("#sort_number").val();
    if (!isNotEmpty(workspace_name)) {
        swal('Warning', '工作空间名称不能为空', 'error');
        return;
    }
    if (!isNotEmpty(sort_number)) {
        swal('Warning', '排序号不能为空', 'error');
        return;
    }
    if (!isNotEmpty(status)) {
        swal('Warning', '请指定工作空间状态', 'error');
        return;
    }
    $.ajax({
        type: 'post',
        url: '/workspace-save',
        data: {
            id: $('#hidden_id').val(),
            workspace_name: workspace_name,
            sort_number: sort_number,
            status: status,
            workspace_description: $("#workspace_description").val(),
            notify: $('#select_notify').val().join(','),
        },
        dataType: 'json',
        success: function (data) {
            show_response_message("编辑工作空间", data, function () {
                location.href = "/workspace-list";
            })
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}


function init_notify_multiselect() {
    $('#select_notify').multiselect({
        nonSelectedText: '选择通知对象',
        enableFiltering: true,
        enableCaseInsensitiveFiltering: true,
        includeFilterClearBtn: true,
        buttonWidth: '100%',
        maxHeight: 600,
        enableClickableOptGroups: true,
        enableCollapsibleOptGroups: true,
        filterPlaceholder: '搜索...',
    });
}

function load_notify_data(selectedData) {
    $.ajax({
        url: '/notify-list-select', // 替换为你的 API 地址
        method: 'POST',
        dataType: 'json',
        success: function (data) {
            // 清空现有的选项
            $('#select_notify').empty();
            // 遍历 JSON 数据，动态生成 <option> 元素
            $.each(data, function (index, item) {
                var option = $('<option></option>')
                    .attr('value', item.id)
                    .text(item.name);
                $('#select_notify').append(option);
            });
            // 初始化或刷新 Multiselect
            if (selectedData) {
                $('#select_notify').val(selectedData);
            }
            $('#select_notify').multiselect('rebuild');
        },
        error: function (xhr, status, error) {
            console.error('加载选项失败:', error);
        }
    });
}