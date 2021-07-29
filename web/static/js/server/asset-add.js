$(function(){
    $("#asset_add").click(function(){
        const org_name = $("#org_name").val();
        const asset_list = $("#asset_list").val();
        const asset_type = $("#asset_type").val();
        if(!org_name || !asset_list){
            swal('Warning', '所属组织和资产列表不能为空', 'error');
        }else{
        $.post("/asset-add",
        {
            "org_name": org_name,
            "asset_list": asset_list,
            "asset_type": asset_type
        },function(data,e){
            if(e === "success"){
                swal({
                    title: "添加资产成功",
                    text: "",
                    type:"success",
                    confirmButtonText: "ok",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: false
                },
                function(){
                    location.href = "/ip-asset-list"
                });
            }else{
                swal('Warning', "添加资产失败!", 'error');
            }
        });}
    });

});