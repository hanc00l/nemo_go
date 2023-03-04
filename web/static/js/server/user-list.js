$(function () {
    $('#btnsiderbar').click();
    $('#user_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/user-list",
                "type": "post",
            },
            columns: [
                {
                    data: "id",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        return '<input type="checkbox" class="checkchild" value="' + data + '"/>';
                    }
                },
                {data: "index", title: "序号", width: "5%"},
                {
                    data: "user_name",
                    title: "用户名称", width: "15%",
                    "render": function (data, type, row, meta) {
                        return '<a onclick="edit_user(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Edit" data-target="#edituser"><span> ' + data + '</span></a>';
                    }
                },
                {data: "user_role", title: "用户类型", width: "8%"},
                {data: "user_description", title: "用户描述", width: "15%"},
                {data: "state", title: "状态", width: "5%"},
                {data: "sort_order", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "12%"},
                {data: "update_time", title: "更新时间", width: "12%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        const strDelete = '<a onclick="delete_user(' + row.id + ')" href="#"><i class="fa fa-trash"></i><span>Delete</span></a>';
                        const strResetPassword = '<a onclick="reset_user_password(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Reset" data-target="#resetpassword"><i class="fa fa-unlock-alt"></i><span>Reset</span></a>';
                        const strUserWorkspace = '<a onclick="set_user_workspace(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Workspace" data-target="#setuserworkspace"><i class="fa fa-cube"></i><span>Workspace</span></a>';

                        return strUserWorkspace + "&nbsp;" + strResetPassword + "&nbsp;" + strDelete;
                    }
                }
            ]
        }
    );//end datatable


    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

    $("#user_update").click(function () {
        const user_id = $("#user_id").val();
        const user_name = $("#user_name").val();
        const user_description = $("#user_description").val();
        const state = $("#state").val();
        const sort_order = $("#sort_order").val();
        const user_role = $("#user_role").val();
        if (!user_id) return;
        if (!user_name) {
            swal('Warning', '用户名称不能为空', 'error');
            return;
        }
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.post("/user-update?id=" + user_id,
            {
                "user_name": user_name,
                "sort_order": sort_order,
                'state': state,
                'user_description': user_description,
                'user_role': user_role,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "更新用户成功",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $("#user_table").DataTable().draw(false);
                            $('#edituser').modal('hide');
                        });
                } else {
                    swal('Warning', "更新用户失败!", 'error');
                }
            });
    });

    $("#reset_user_password").click(function () {
        const user_id = $("#reset_user_id").val();
        const user_password1 = $("#reset_user_password1").val();
        const user_password2 = $("#reset_user_password2").val();
        if (!user_id || !user_password1 || !user_password2) return;
        if (user_password1 !== user_password2) {
            swal('Warning', '两次密码不一致！', 'error');
            return;
        }
        $.post("/user-reset-password",
            {
                'id': user_id,
                'user_password': user_password1,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "重置密码成功！",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $('#resetpassword').modal('hide');
                        });
                } else {
                    swal('Warning', "重置密码失败!", 'error');
                }
            });
    });

    $("#set_user_workspace").click(function () {
        const user_id = $("#userworkspace_user_id").val();
        if (!user_id) return;

        let selected_workspace_id = "";
        const checkSelected = $('input:checkbox[name="cb_userworkspace"]');
        for (let i = 0; i < checkSelected.length; i++) {
            if (checkSelected[i].checked) {
                if (checkSelected[i].value !== '') {
                    if (selected_workspace_id === "") selected_workspace_id = checkSelected[i].value;
                    else selected_workspace_id += ',' + checkSelected[i].value;
                }
            }
        }
        $.post("/user-workspace-update",
            {
                'user_id': user_id,
                'workspace_id': selected_workspace_id,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "保存权限成功！",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $('#setuserworkspace').modal('hide');
                        });
                } else {
                    swal('Warning', "保存权限失败!", 'error');
                }
            });
    });
});

function delete_user(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除该用户的信息！",
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
                url: '/user-delete?id=' + id,
                success: function (data) {
                    $("#user_table").DataTable().draw(false);
                },
                error: function (xhr, type) {
                }
            });
        });
}


function edit_user(id) {
    $.ajax({
        type: 'post',
        url: '/user-get?id=' + id,
        dataType: 'json',
        success: function (e) {
            const data = eval(e);
            $('#user_name').val(data.user_name);
            $('#sort_order').val(data.sort_order);
            $('#state').val(data.state);
            $('#user_id').val(id);
            $('#user_description').val(data.user_description);
            $('#user_role').val(data.user_role);
        },
        error: function (xhr, type) {
        }
    });
}

function reset_user_password(id) {
    $('#reset_user_id').val(id);
}

function set_user_workspace(id) {
    $('#userworkspace_user_id').val(id);
    $('#checkbox_user_workspace_list').empty();
    $.post("/user-workspace-list",
        {
            "user_id": id
        }, function (data, e) {
            if (e === "success") {
                let strCheckBox = '';
                for (let i = 0; i < data.length; i++) {
                    strCheckBox += '<div className="form-check"><input className="form-check-input" type="checkbox" value="';
                    strCheckBox += data[i].workspaceId;
                    strCheckBox += '" name="cb_userworkspace" id="cb_userworkspace_';
                    strCheckBox += data[i].workspaceId;
                    strCheckBox += '"';
                    if (data[i].enable === true) strCheckBox += ' checked';
                    strCheckBox += '><label className="form-check-label" for="cb_userworkspace_';
                    strCheckBox += data[i].workspaceId;
                    strCheckBox += '">&nbsp;';
                    strCheckBox += data[i].workspaceName;
                    strCheckBox += '</label></div>';
                }
                $('#checkbox_user_workspace_list').append(strCheckBox);
            }
        });
}