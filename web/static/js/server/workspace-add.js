$(function () {
    $("#workspace_add").click(function () {
        const workspace_name = $("#workspace_name").val();
        const state = $("#state").val();
        const workspace_description = $("#workspace_description").val();
        const sort_order = $("#sort_order").val();
        if (!workspace_name) {
            swal('Warning', '工作空间名称不能为空', 'error');
            return;
        }
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        if (!state) {
            swal('Warning', '请指定工作空间状态', 'error');
            return;
        }
        $.post("/workspace-add",
            {
                "workspace_name": workspace_name,
                "sort_order": sort_order,
                'state': state,
                'workspace_description': workspace_description,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "添加工作空间成功",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            location.href = "/workspace-list"
                        });
                } else {
                    swal('Warning', "添加工作空间失败!", 'error');
                }
            });

    });
});