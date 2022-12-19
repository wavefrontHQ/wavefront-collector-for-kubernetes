require './dashboards_validator.rb'

class FakeDashboardIterator
end

describe ColorChecker do
  describe '.line_with_closest_rgba' do
    it 'returns the RGBA value closest to a valid one in the Palette' do
      cc = ColorChecker.new FakeDashboardIterator.new
      expect(cc.line_with_closest_rgba('                "sparklineLineColor": "rgba(0,0,0,0.3)",', 'rgba(0,0,0,0.3)'))
                                          .to eq('                "sparklineLineColor": "rgba(0,0,0,1)",')
      expect(cc.line_with_closest_rgba('                "sparklineLineColor": "rgba(255,50,50,1)",', 'rgba(255,50,50,1)'))
                                          .to eq('                "sparklineLineColor": "rgba(245,79,71,1)",')
      expect(cc.line_with_closest_rgba('                "sparklineLineColor": "rgba(0,255,0,1)",', 'rgba(0,255,0,1)'))
                                          .to eq('                "sparklineLineColor": "rgba(133,200,26,1)",')
      expect(cc.line_with_closest_rgba('                "sparklineLineColor": "rgba(50,50,255,1)",', 'rgba(50,50,255,1)'))
                                          .to eq('                "sparklineLineColor": "rgba(73,175,217,1)",')
    end
  end
end