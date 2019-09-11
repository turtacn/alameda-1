package stats

import "testing"

func TestMAPE(t *testing.T) {
	type args struct {
		measurementDataSet map[int64]*MeasurementData
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				measurementDataSet: map[int64]*MeasurementData{
					1: &MeasurementData{
						predictData: data{
							value: 90,
						},
						metricData: data{
							value: 100,
						},
					},
					2: &MeasurementData{
						predictData: data{
							value: 100,
						},
						metricData: data{
							value: 80,
						},
					},
					3: &MeasurementData{
						predictData: data{
							value: 150,
						},
						metricData: data{
							value: 100,
						},
					},
					4: &MeasurementData{
						predictData: data{
							value: 100,
						},
						metricData: data{
							value: 200,
						},
					},
					5: &MeasurementData{
						predictData: data{
							value: 100,
						},
						metricData: data{
							value: 50,
						},
					},
				}},
			want: 47,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MAPE(tt.args.measurementDataSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("MAPE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MAPE() = %v, want %v", got, tt.want)
			}
		})
	}
}
