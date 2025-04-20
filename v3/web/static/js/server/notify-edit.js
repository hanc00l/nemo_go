$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    const notify_id = $.urlParam('id');
    if (isNotEmpty(notify_id)) {
        load_notify(notify_id);
    }
    $('#notify_form').on('submit', function (e) {
        e.preventDefault();
        save_notify_data();
    });
    $('#btn_cancel').click(function () {
        history.back();
    });
});

function save_notify_data() {
    const notify_name = $("#notify_name").val();
    const status = $("#status").val();
    const notify_category = $("#notify_category").val();
    const notify_description = $("#notify_description").val();
    const sort_number = $("#sort_number").val();
    const notify_token = $("#notify_token").val();
    const notify_template = $("#notify_template").val();

    if (!isNotEmpty(notify_name)) {
        swal('Warning', '通知名称不能为空', 'error');
        return;
    }
    if (!isNotEmpty(notify_category)) {
        swal('Warning', '通知分类不能为空', 'error');
        return;
    }
    if (!isNotEmpty(sort_number)) {
        swal('Warning', '排序号不能为空', 'error');
        return;
    }
    if (!isNotEmpty(status)) {
        swal('Warning', '请指定通知状态', 'error');
        return;
    }
    if (!isNotEmpty(notify_token)) {
        swal('Warning', '通知token不能为空', 'error');
        return;
    }
    if (!isNotEmpty(notify_template)) {
        swal('Warning', '通知模板不能为空', 'error');
        return;
    }
    const notify_id = $.urlParam('id');
    let url = '/notify-add';
    if (isNotEmpty(notify_id)) {
        url = `/notify-update?id=${notify_id}`;
    }
    $.ajax({
        type: 'post',
        url: url,
        data: {
            name: notify_name,
            category: notify_category,
            token: notify_token,
            template: notify_template,
            sort_number: sort_number,
            status: status,
            description: notify_description,
        },
        dataType: 'json',
        success: function (data) {
            show_response_message("保存通知", data, function () {
                location.href = "/notify-list"
            });
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}

function load_notify(id) {
    $.ajax({
        type: 'post',
        url: '/notify-get',
        data: {id: id},
        dataType: 'json',
        success: function (data) {
            $('#notify_name').val(data.name);
            $('#notify_token').val(data.token);
            $('#notify_template').val(data.template);
            $('#sort_number').val(data.sort_number);
            $('#status').val(data.status);
            $('#notify_description').val(data.description);
            $('#notify_category').val(data.category);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}

