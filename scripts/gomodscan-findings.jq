# Fail closed on malformed reports or waivers before matching any finding.
def nonempty: type == "string" and test("\\S");
def key: [.module, .version, .finding];

(if ($accepted | length != 1) or ($accepted[0] | type != "array") then
  error("gomodscan: expected one accepted-findings array")
else $accepted[0] end) as $waivers
| if all($waivers[];
    type == "object" and (keys == ["finding", "module", "reason", "version"]) and
    (.module | nonempty) and (.version | nonempty) and
    (.finding | nonempty) and (.reason | nonempty)) and
    (($waivers | map(key) | unique | length) == ($waivers | length)) then .
  else error("gomodscan: invalid or duplicate accepted finding; each entry needs a reason") end
| if length == 1 and
    all(.[0].runs[]; .tool.driver.name == "gomodscan" and
      all(.results[]; (.ruleId | nonempty) and
        (.locations[0].logicalLocations[0].fullyQualifiedName | nonempty) and
        (.locations[0].logicalLocations[0].kind == "module"))) then .[0]
  else error("gomodscan: unexpected finding format") end
| [.runs[].results[] |
    (.locations[0].logicalLocations[0].fullyQualifiedName | split("@")) as $name
    | if ($name | length) != 2 or any($name[]; nonempty | not) then
        error("gomodscan: invalid module and version")
      else {module: $name[0], version: $name[1],
        finding: (if .ruleId == "malicious-package" then
          if (.properties.id | nonempty) then .ruleId + "/" + .properties.id
          else error("gomodscan: missing malicious-package identifier") end
          else .ruleId end)} end] as $findings
| [$findings[] | . as $finding |
    [$waivers[] | select(key == ($finding | key))] as $matches
    | if ($matches | length) == 1 then
        "gomodscan: accepted \(.finding): \(.module)@\(.version); reason: \($matches[0].reason)"
      else "gomodscan: unaccepted \(.finding): \(.module)@\(.version)" end] as $messages
| [$waivers[] | . as $waiver |
    select(all($findings[]; key != ($waiver | key))) |
    "gomodscan: stale accepted finding \(.finding): \(.module)@\(.version)"] as $stale
| ($messages + $stale)[],
  (if ($stale | length) > 0 or any($messages[]; startswith("gomodscan: unaccepted")) then
    error("gomodscan: findings do not exactly match the accepted file")
   else "gomodscan: all findings matched; no stale accepted entries" end)
