$(function () {
    $("#org_add").click(function () {
        const org_name = $("#org_name").val();
        const status = $("#status").val();
        const sort_order = $("#sort_order").val();
        if (!org_name) {
            swal('Warning', '组织名称不能为空', 'error');
            return;
        } 
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.post("/org-add",
            {
                "org_name": org_name,
                "sort_order": sort_order,
                'status': status
            }, function (data, e) {
                if (e === "success") {
                    swal({
                        title: "添加组织成功",
                        text: "",
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    },
                        function () {
                            location.href = "/org-list"
                        });
                } else {
                    swal('Warning', "添加机构失败!", 'error');
                }
            });

    });
});