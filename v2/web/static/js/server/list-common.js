/**
 * 判断是否是IP地址
 * @type {function(*=): (boolean|*)}
 */
let isIpv4 = function () {
    var regexp = /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/;
    return function (value) {
        var valid = regexp.test(value);
        if (!valid) {//首先必须是 xxx.xxx.xxx.xxx 类型的数字，如果不是，返回false
            return false;
        }
        return value.split('.').every(function (num) {
            //切割开来，每个都做对比，可以为0，可以小于等于255，但是不可以0开头的俩位数
            //只要有一个不符合就返回false
            if (num.length > 1 && num.charAt(0) === '0') {
                //大于1位的，开头都不可以是‘0’
                return false;
            } else if (parseInt(num, 10) > 255) {
                //大于255的不能通过
                return false;
            }
            return true;
        });
    }
}();

/**
 * 判断是否是IPV6地址
 * @param str
 * @returns {boolean}
 */
let isIpv6 = function (str) {
    var pattern = /(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))/;
    return pattern.test(str);
}

/**
 * 将缩写的IPv6地址转换为完整的IPv6地址
 * @param simpeIpv6
 * @returns {string}
 */
let tranSimIpv6ToFullIpv6 = function (simpeIpv6) {
    simpeIpv6 = simpeIpv6.toUpperCase()
    if (simpeIpv6 == "::") {
        return "0000:0000:0000:0000:0000:0000:0000:0000";
    }
    let arr = ["0000", "0000", "0000", "0000", "0000", "0000", "0000", "0000"]
    if (simpeIpv6.startsWith("::")) {
        let tmpArr = simpeIpv6.substring(2).split(":")
        for (let i = 0; i < tmpArr.length; i++) {
            arr[i + 8 - tmpArr.length] = ('0000' + tmpArr[i]).slice(-4)
        }
    } else if (simpeIpv6.endsWith("::")) {
        let tmpArr = simpeIpv6.substring(0, simpeIpv6.length - 2).split(":");
        for (let i = 0; i < tmpArr.length; i++) {
            arr[i] = ('0000' + tmpArr[i]).slice(-4)
        }
    } else if (simpeIpv6.indexOf("::") >= 0) {
        let tmpArr = simpeIpv6.split("::");
        let tmpArr0 = tmpArr[0].split(":");
        for (let i = 0; i < tmpArr0.length; i++) {
            arr[i] = ('0000' + tmpArr0[i]).slice(-4)
        }
        let tmpArr1 = tmpArr[1].split(":");
        for (let i = 0; i < tmpArr1.length; i++) {
            arr[i + 8 - tmpArr1.length] = ('0000' + tmpArr1[i]).slice(-4)
        }
    } else {
        let tmpArr = simpeIpv6.split(":");
        for (let i = 0; i < tmpArr.length; i++) {
            arr[i + 8 - tmpArr.length] = ('0000' + tmpArr[i]).slice(-4)
        }
    }
    return arr.join(":")
}

/**
 * 获取完整的IPv6地址
 * @param ip_string
 * @returns {*}
 */
let getFullIpv6 = function (ip_string) {
    // take care of leading and trailing ::
    ip_string = ip_string.replace(/^:|:$/g, '');

    const ipv6 = ip_string.split(':');

    for (let i = 0; i < ipv6.length; i++) {
        let hex = ipv6[i];
        if (hex != "") {
            // normalize leading zeros
            ipv6[i] = ("0000" + hex).substr(-4);
        } else {
            // normalize grouped zeros ::
            hex = [];
            for (let j = ipv6.length; j <= 8; j++) {
                hex.push('0000');
            }
            ipv6[i] = hex.join(':');
        }
    }

    return ipv6.join(':');
}

/**
 * 将完整的IPv6地址转换为缩写的IPv6地址
 * @param ipv6
 */
let compressIPv6 = function (ip) {
    //First remove the leading 0s of the octets. If it's '0000', replace with '0'
    let output = ip.split(':').map(terms => terms.replace(/\b0+/g, '') || '0').join(":");

    //Then search for all occurrences of continuous '0' octets
    let zeros = [...output.matchAll(/\b:?(?:0+:?){2,}/g)];

    //If there are occurences, see which is the longest one and replace it with '::'
    if (zeros.length > 0) {
        let max = '';
        zeros.forEach(item => {
            if (item[0].replaceAll(':', '').length > max.replaceAll(':', '').length) {
                max = item[0];
            }
        })
        output = output.replace(max, '::');
    }
    return output;
}

/**
 * 获取IPv6地址的C段地址
 * @param ipv6
 * @returns {string}
 */
let getIPv6CSubnet = function (ipv6) {
    let ipv6Full = getFullIpv6(ipv6)
    let ipv6Arr = ipv6Full.split(":")
    let ipv6CSubnet = ipv6Arr[0] + ":" + ipv6Arr[1] + ":" + ipv6Arr[2] + ":" + ipv6Arr[3] + ":" + ipv6Arr[4] + ":" + ipv6Arr[5] + ":" + ipv6Arr[6] + ":" + "0000"
    let ipv6CSubnetComp = compressIPv6(ipv6CSubnet)
    return ipv6CSubnetComp + "/120"
}