require './dashboards_validator.rb'

class FakeDashboardIterator
end

describe ColorChecker do
  describe '.line_with_closest_rgba' do
    it 'returns the RGBA value closest to a valid one in the Palette, preserving spacing' do
      cc = ColorChecker.new FakeDashboardIterator.new

      # Colors
      expect(cc.line_with_closest_rgba('                "sparklineLineColor": "rgba(0,0,0,0.3)",', 'rgba(0,0,0,0.3)'))
                                .to eq('                "sparklineLineColor": "rgba(0,0,0,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(255,50,50,1)",', 'rgba(255,50,50,1)'))
                                .to eq('"sparklineLineColor": "rgba(245,79,71,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(0,255,0,1)",', 'rgba(0,255,0,1)'))
                                .to eq('"sparklineLineColor": "rgba(133,200,26,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(50,50,255,1)",', 'rgba(50,50,255,1)'))
                                .to eq('"sparklineLineColor": "rgba(73,175,217,1)",')

      # Black and white
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(200,200,200,1)",', 'rgba(200,200,200,1)'))
                                .to eq('"sparklineLineColor": "rgba(255,255,255,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(50,50,50,1)",', 'rgba(50,50,50,1)'))
                                .to eq('"sparklineLineColor": "rgba(0,0,0,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(200,200,200,0.5)",', 'rgba(200,200,200,0.5)'))
                                .to eq('"sparklineLineColor": "rgba(255,255,255,1)",')
      expect(cc.line_with_closest_rgba('"sparklineLineColor": "rgba(50,50,50,0.5)",', 'rgba(50,50,50,0.5)'))
                                .to eq('"sparklineLineColor": "rgba(0,0,0,1)",')
    end
  end

  describe '.fix_hex_and_rgba' do
    it 'finds and fixes either hex or rgba' do
      cc = ColorChecker.new FakeDashboardIterator.new

      expect(cc.fix_hex_and_rgba('"sparklineDisplayColor": "rgba(250,250,250,0.9)",'))
                          .to eq('"sparklineDisplayColor": "rgba(255,255,255,1)",')
      expect(cc.fix_hex_and_rgba('"colors": ["#ffbf00"]'))
                          .to eq('"colors": ["#f57600"]')
    end

    it 'should fix raw r, g, or b, not convert it to closest other color' do
      # TODO please implement
    end
  end
end
