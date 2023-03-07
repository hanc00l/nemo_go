$(function () {
    $("#user_add").click(function () {
        const user_name = $("#user_name").val();
        const state = $("#state").val();
        const user_role = $("#user_role").val();
        const user_description = $("#user_description").val();
        const sort_order = $("#sort_order").val();
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
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        if (!state) {
            swal('Warning', '请指定用户状态', 'error');
            return;
        }
        $.post("/user-add",
            {
                "user_name": user_name,
                "sort_order": sort_order,
                'state': state,
                'user_description': user_description,
                'user_role': user_role,
                'user_password': user_password,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "添加用户成功",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            location.href = "/user-list"
                        });
                } else {
                    swal('Warning', "添加用户失败!", 'error');
                }
            });

    });
});