$(function () {
    $("#user_add").click(function () {
        const user_name = $("#user_name").val();
        const status = $("#status").val();
        const user_role = $("#user_role").val();
        const user_description = $("#user_description").val();
        const sort_number = $("#sort_number").val();
        const user_password = $("#user_password").val();
        const user_password_confirm = $("#user_password_confirm").val();
        if (!user_name) {
            swal('Warning', '用户名称不能为空', 'error');
            return;
        }
        if (!user_password || !user_password_confirm) {
            swal('Warning', '登录密码不能为空', 'error');
            return;
        }
        if (user_password !== user_password_confirm) {
            swal('Warning', '两次密码不一致', 'error');
            return;
        }
        if (!sort_number) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        if (!status) {
            swal('Warning', '请指定用户状态', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/user-add',
            data: {
                user_name: user_name,
                sort_number: sort_number,
                status: status,
                user_description: user_description,
                user_role: user_role,
                user_password: user_password,
            },
            success: function (data) {
                show_response_message("新增用户", data, function () {
                    location.href = "/user-list"
                });
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
});
