<% wrap_layout :inner do %>
  <% content_for :sidebar do %>
    <div class="docs-sidebar hidden-print affix-top" role="complementary">
      <ul class="nav docs-sidenav">
        <li<%= sidebar_current("docs-home") %>>
            <a href="/docs/providers/index.html">All Providers</a>
        </li>

        <li<%= sidebar_current("docs-elkaliases-index") %>>
            <a href="/docs/providers/elkaliases/idex.html">ElkAliases Provider</a>
        </li>

        <li<%= sidebar_current("docs-elkaliases-resource") %>>
        <a href="#">Resources</a>
            <ul class="nav nav-visible">
                <li<%= sidebar_current("docs-elkaliases-resource-elkaliases_index") %>>
                    <a href="/docs/providers/elkaliases/r/elkaliases_index.html">elkaliases_index</a>
                </li>
            </ul>
        </li>
      </ul>
    </div>
  <% end %>

  <%= yield %>
  <% end %>
