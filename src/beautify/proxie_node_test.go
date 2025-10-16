package beautify

import "testing"

func Test_getRate(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// 基本倍率在前缀中
		{
			name: "brackets 0.25x",
			args: args{name: "[0.25x]🇨🇳 测试节点一"},
			want: "[0.25x]",
		},
		{
			name: "brackets 1.0x",
			args: args{name: "[1.0x]🇺🇦 测试节点二"},
			want: "",
		},
		{
			name: "mixed bracket in name",
			args: args{name: "[1.0xUK 英国|城域专线"},
			want: "",
		},
		// 括号漏掉或者格式变异
		{
			name: "inline bracket no end symbol",
			args: args{name: "[0.50x🇨🇯 澳大利亚|实验"},
			want: "[0.50x]",
		},
		// 倍率在末尾
		{
			name: "suffix rate 0.5x",
			args: args{name: "🇨🇦加拿大·中部 |0.5x 1"},
			want: "[0.5x]",
		},
		{
			name: "suffix rate 0.5x with extra label",
			args: args{name: "🇨🇦加拿大·中部 |0.5x 场景2"},
			want: "[0.5x]",
		},
		// 没有倍率
		{
			name: "line no rate 1",
			args: args{name: "剩余流量：123.45 GB"},
			want: "",
		},
		{
			name: "line no rate 2",
			args: args{name: "距离下次重置剩余：11 天"},
			want: "",
		},
		{
			name: "expire date line",
			args: args{name: "套餐到期：2299-02-01"},
			want: "",
		},
		// 倍率带特殊字符或空格
		{
			name: "rate at middle with pipe",
			args: args{name: "🇯🇵日本|福利家宽softbank|动态ip|1.5x"},
			want: "[1.5x]",
		},
		{
			name: "rate at middle with space",
			args: args{name: "节点节点|测试 2.2x 专线"},
			want: "[2.2x]",
		},
		{
			name: "rate at start without bracket",
			args: args{name: "0.8x-日本|分流"},
			want: "[0.8x]",
		},
		// 畸形/不标准倍率
		{
			name: "invalid rate text",
			args: args{name: "美国节点-普通"},
			want: "",
		},
		{
			name: "broken x multiplier",
			args: args{name: "0.x日本|家宽"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRate(tt.args.name); got != tt.want {
				t.Errorf("getRate() = %v, want %v", got, tt.want)
			}
		})
	}
}
