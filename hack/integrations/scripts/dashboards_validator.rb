require 'json'

class Palette
  BACKGROUND_COLORS = [
    "#f54f47", # Red
    "#f57600", # Orange
    "#85c81a", # Green
    "#49afd9", # neutral Blue
  ]

  OTHER_COLORS = [
    "#000000", # Pure White
    "#ffffff", # Pure Black
  ]

  def self.all_colors
    BACKGROUND_COLORS + OTHER_COLORS
  end

  def self.print
    puts "Palette"
    self.all_colors.sort.each do |color_string|
      color = Color.from_string(color_string)
      printf("\x1b[48;2;#{color.r};#{color.g};#{color.b}m #{" " * 20}")
      printf("\033[0m")

      puts(color_string + "\t" + color.rgba)
    end
  end
end

def rgba(hex_string)
  colors = hex_string.gsub("#", "").chars.each_slice(2).map do |hex|
    hex.join.to_i(16)
  end.to_a
  "rgba(#{colors.join(",")},1)"
end

class Color
  attr_reader :r, :g, :b

  def initialize(r, g, b)
    @r = r
    @g = g
    @b = b
  end

  def self.from_string(str)
    HEX_PATTERN.match?(str) ?
      from_hex(str) : from_rgba(str)
  end

  def self.from_hex(hex_string)
    args = hex_string.gsub("#", "").chars.each_slice(2).map do |hex|
      hex.join.to_i(16)
    end.to_a
    new(*args)
  end

  def self.from_rgba(rgba_string)
    #rgba(255,0,0,1)
    m = /rgba\((?<r>\d+),(?<g>\d+),(?<b>\d+)/.match(rgba_string.gsub(" ", ""))
    new(m[:r], m[:g], m[:b])
  end

  def rgba
    "rgba(#{r},#{g},#{b},1)"
  end
end

NonMatch = Struct.new(:dashboard_file, :lineno, :line, :match) do
  def to_s
    "#{dashboard_file}:#{lineno}: #{line.gsub("\n", '')}"
  end
end

HEX_PATTERN = /#[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]/
RGBA_PATTERN = /rgb.*\)/

class ProseFormatting
  NON_CAPS_WORDS = %(of the vs to by per in with no for over a but from and v1 v2 on)

  def self.ends_with_period?(str)
    str[-1] == "."
  end

  def self.title_case?(str)
    if str.nil? || str == ""
      return true
    end
    str.split.all? { |w| capitalized?(w) || NON_CAPS_WORDS.include?(w.gsub(/[.()]/, '')) }
  end

  def self.capitalized?(str)
    str[0] == str[0].capitalize
  end

  def self.paragraphs?(str)
    str.split(".").size > 1 || str.include?(",")
  end
end

class ColorChecker
  def initialize(iterator)
    @iterator = iterator
    @non_matches = []
  end

  def run(reporter)
    @iterator.dashboard_files.each do |file_path|
      File.open(file_path) do |file|
        file.each_line do |line|
          check_hex(file_path, file, line)
          check_rgba(file_path, file, line)
        end
      end
    end

    @non_matches.each do |m|
      reporter.report(m)
    end

    print_results
  end

  def print_results
    @non_matches.each do |nonmatch|
      color = Color.from_string(nonmatch.match)
      printf("\x1b[48;2;#{color.r};#{color.g};#{color.b}m #{" " * 20}")
      printf("\033[0m")
      puts(nonmatch.to_s)
    end

    puts "--- Summary ---"
    puts "#{@non_matches.empty? ? "✅" : "❌" } Non-palette colors used: #{@non_matches.size}"
  end

  private

  def check_hex(dashboard_file, file, line)
    begin
      HEX_PATTERN.match(line).tap do |match|
        if match
          unless matches_any_hex?(match.to_s)
            record(dashboard_file, file.lineno, line, match.to_s)
          end
        end
      end
    rescue
      puts 'Encountered an error matching regex, assuming this was not a color'
    end
  end

  def record(dashboard_file, lineno, fulline, match)
    @non_matches << NonMatch.new(dashboard_file, lineno, fulline, match)
  end

  def check_rgba(dashboard_file, file, line)
    begin
      RGBA_PATTERN.match(line).tap do |match|
        if match
          unless matches_any_rgba?(match.to_s)
            record(dashboard_file, file.lineno, line, match.to_s)
          end
        end
      end
    rescue
      puts 'Encountered an error matching regex, assuming this was not a color'
    end
  end

  def matches_any_hex?(hex_color)
    !all_colors.select { |color| !!color.casecmp?(hex_color) }.none?
  end

  def matches_any_rgba?(rgba_color)
    !all_colors.map { |hex| rgba(hex) }.select { |color| !!color.casecmp?(rgba_color) }.none?
  end

  def all_colors
    Palette.all_colors
  end
end

class ChartTitleChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def exception?(str)
    str == "Make your own version..."
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      title = chart["name"]
      unless ProseFormatting.title_case?(title) || exception?(title)
        reporter.report(Reporter::Issue.new("Chart Title not in Title Case", title, dashboard_name))
      end
    end
  end
end

class SparklineSublabelChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def exception?(str)
    str == "of Router Jobs Running"
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      if chart["chartAttributes"] && chart["chartAttributes"]["singleStat"]
        sparkline_label = chart["chartAttributes"]["singleStat"]["sparklineDisplaySubLabel"]
        unless ProseFormatting.title_case?(sparkline_label) || exception?(sparkline_label)
          reporter.report(Reporter::Issue.new("sparklineDisplaySubLabel not in Title Case", sparkline_label, dashboard_name))
        end
      end
    end
  end

end

class SparklineFontSizeChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      if chart["chartSettings"] && chart["chartSettings"]["sparklineDisplayFontSize"]
        font_size = chart["chartSettings"]["sparklineDisplayFontSize"]
        unless font_size == "150"
          reporter.report(Reporter::Issue.new("sparklineDisplayFontSize not 150", font_size, dashboard_name))
        end
      end
    end
  end

end

class ChartDescriptionChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def exception?(str)
    [
      "Number of Deployments in the processing state of type update or delete deployment.",
      "Number of deployments in the queued state of type update or delete deployment."
    ].include?(str)
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      description = chart["description"]
      if chart["chartSettings"] && (chart["chartSettings"]["type"] == "markdown-widget")
        next
      end
      if (description == "" or description.nil?)
        reporter.report(Reporter::Issue.new("Description should be empty", chart["name"], dashboard_name))
        next
      end
      unless ProseFormatting.capitalized?(description) && (ProseFormatting.ends_with_period?(description) || ProseFormatting.paragraphs?(description))
        reporter.report(Reporter::Issue.new("Description not in Title Case", description, dashboard_name))
      end
    end
  end
end

class DashboardLinkChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      if chart["chartAttributes"] && chart["chartAttributes"]["dashboardLinks"]
        links = chart["chartAttributes"]["dashboardLinks"]
        links.each do |key, link|
          unless link["destination"].start_with?("/dashboards/integration-tas-v4")
            reporter.report(Reporter::Issue.new("dashboard link destination not canonical url", link["destination"], dashboard_name))
          end
        end
      end
    end
  end
end

class DashboardIterator
  def initialize(file_glob)
    @file_glob = file_glob
  end

  def dashboard_files
    Dir.glob(@file_glob)
  end

  def each_dashboard
    if dashboard_files.size == 0
      raise "No dashboard files found"
    end
    dashboard_files.each do |file|
      yield JSON.load(File.read(file))
    end
  end

  def each_chart
    each_dashboard do |dashboard_json|
      dashboard_json["sections"].each do |section|
        section["rows"].each do |row|
          row["charts"].each do |chart|
            yield chart, dashboard_json["name"]
          end
        end
      end
    end
  end
end

class ChartUnitChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def valid_unit?(item)
    [
      'Messages',
      "%",
      "#",
      "Markdown",
      '1 = Enabled, 0 = Disabled',
      '1 = Healthy, 0 = Unhealthy',
      '1 = Success, 0 = Failure',
      'Chunks',
      'Crashes',
      'Bytes',
      'Hits',
      'Misses',
      'ms',
      'Errors',
      'Metrics',
      'Requests',
      'Clients',
      'Queries per Second',
      'Cache Hits',
      'Threads',
      'Tables',
      'Envelopes',
      'Connections per Second',
      'Connections',
      'Commands per Second',
      'Queries',
      'GiB',
      'MiB',
      'ns',
      'B',
      's',
      'Seconds',
      'Failures per Second',
      'millicores',
      'bps',
      'pps',
      'items'
    ].include?(item)
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      unit = chart["units"]
      next if unit.nil?

      if chart["chartSettings"] && (chart["chartSettings"]["type"] == "markdown-widget")
        unless unit == ""
          reporter.report(Reporter::Issue.new("Markdown charts should not have a unit defined", unit, "#{dashboard_name}: #{chart["name"]}"))
        end
         next
      end

     if chart["chartSettings"] && (chart["chartSettings"]["showValueColumn"] == false)
       unless unit == ""
         reporter.report(Reporter::Issue.new("Charts with no value column should not have a unit defined", unit, "#{dashboard_name}: #{chart["name"]}"))
       end
        next
     end

      if chart["chartSettings"] && ["sparkline", "gauge"].include?(chart["chartSettings"]["type"])
        unless valid_sparkline_unit?(unit)
          reporter.report(Reporter::Issue.new("Unrecognized sparkline/gauge chart unit", unit, "#{dashboard_name}: #{chart["name"]}"))
        end
        next
      end

      unless valid_unit?(unit)
        reporter.report(Reporter::Issue.new("Unrecognized unit", unit, "#{dashboard_name}: #{chart["name"]}"))
      end
    end
  end

  private

  def valid_sparkline_unit?(unit)
    ["%", "", "B", "ms", "bps"].include? unit
  end
end

class ChartQueryChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def run(reporter)
    @iterator.each_chart do |chart, dashboard_name|
      chart["sources"].each do |source|
        query = source["query"]

        QueryChecks.unquoted_metrics(query).each do |unquoted_metric|
          reporter.report(Reporter::Issue.new("Unquoted metric name #{unquoted_metric}: ", query, dashboard_name))
        end
        QueryChecks.unquoted_variables(query).each do |unquoted_var|
          reporter.report(Reporter::Issue.new("Unquoted variable in filter expression: #{unquoted_var}", query, dashboard_name))
        end
      end
    end
  end
end

class QueryChecks

  UNQUOTED_VAR=/(?<==)\${[a-z_]*}/
  UNQUOTED_TAS_METRIC_NAME=/(?<!")tas\.(?:\w|\.|-)*/
  UNQUOTED_METRIC_NAME=/(?<=ts\()(?:\w|\.|-|~|\*)+/

  def self.unquoted_metrics(query)
    if query == 'label_replace(tas.gorouter.file_descriptors, "placement_tag", "cf", "placement_tag", "")'
      return []
    end
    query.scan(UNQUOTED_TAS_METRIC_NAME) + query.scan(UNQUOTED_METRIC_NAME)
  end

  def self.unquoted_variables(query)
    query.scan(UNQUOTED_VAR)
  end
end

class ParameterQueryChecker
  def initialize(iterator)
    @iterator = iterator
  end

  def run(reporter)
    @iterator.each_dashboard do |dashboard|
      dashboard_name = dashboard["name"]
      dashboard["parameterDetails"].each do |_, param|
        if param["queryValue"]
          query = param["queryValue"]
          QueryChecks.unquoted_metrics(query).each do |unquoted_metric|
            reporter.report(Reporter::Issue.new("Unquoted metric name #{unquoted_metric}: ", query, dashboard_name))
          end
          QueryChecks.unquoted_variables(query).each do |unquoted_var|
            reporter.report(Reporter::Issue.new("Unquoted variable in filter expression: #{unquoted_var}", query, dashboard_name))
          end
        end
      end
    end
  end
end

class Reporter

  Issue = Struct.new(:summary, :snippet, :dashboard) do
    def to_s
      "#{summary}: #{snippet} (#{dashboard})"
    end
  end

  def initialize
    @issues = []
  end

  def report(issue)
    @issues << issue
  end

  def summarize
    @issues.uniq!

    @issues.each do |issue|
      puts issue
    end

    if @issues.any?
      puts "🙀 Found things to fix 💻"
    else
      puts "✅ No reportable linting issues found 🧡"
    end
  end

  def exit
    if @issues.any?
      Kernel.exit 1
    end
    Kernel.exit 0
  end
end

if ARGV.empty?
  puts "Please enter integration directory. Ex: ../tomcat"
  exit
end

Palette.print
integration_dir_or_file = ARGV[0]
integration_file_glob = integration_dir_or_file.end_with?(".json") ? integration_dir_or_file : "#{integration_dir_or_file}/dashboards/*.json"

dashboards = DashboardIterator.new(integration_file_glob)
reporter = Reporter.new
ColorChecker.new(dashboards).run(reporter)
ChartTitleChecker.new(dashboards).run(reporter)
SparklineSublabelChecker.new(dashboards).run(reporter)
SparklineFontSizeChecker.new(dashboards).run(reporter)
ChartDescriptionChecker.new(dashboards).run(reporter)
ChartUnitChecker.new(dashboards).run(reporter)
DashboardLinkChecker.new(dashboards).run(reporter)
ChartQueryChecker.new(dashboards).run(reporter)
ParameterQueryChecker.new(dashboards).run(reporter)

reporter.summarize
reporter.exit
