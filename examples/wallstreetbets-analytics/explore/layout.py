"""
  Exposes the helper functions
    - `graph_config`
    - `make_layout`
    - `make_double_plot`
    - `plot_and_save`
  which generates Plotly layout objects that adhere to a shared visual standard
"""

from plotly import tools
import plotly.graph_objs as go
import plotly.io as pio
import plotly.offline as plt

""" Background colors """
BACKGROUND_LIGHT = "#343f68"
BACKGROUND_NORMAL = "#1b2442"
BACKGROUND_DARK = "#11182e"
BACKGROUND_VERY_DARK = "#060914"

""" Colors used for traces in plots """
COLORWAY = [
    "rgb(60, 170, 255)",  # blue
    "rgb(255, 60, 130)",  # red
    "rgb(100, 200, 100)",  # green
    "rgb(200, 220, 240)",  # gray
    "rgb(251, 241, 67)",  # yellow
    "rgb(203, 60, 255)",  # purple
    "rgb(249, 148, 66)",  # orange
    "rgb(36, 130, 36)",  # dark green
    "rgb(238, 255, 234)",  # pale green
]

""" Default config for plotly graphs """
graph_config = {
    "displaylogo": False,
    "modeBarButtonsToRemove": [
        "toImage",
        "zoom2d",
        "pan2d",
        "select2d",
        "lasso2d",
        "zoomIn2d",
        "zoomOut2d",
        "autoScale2d",
        "resetScale2d",
        "hoverClosestCartesian",
        "hoverCompareCartesian",
        "zoom3d",
        "pan3d",
        "orbitRotation",
        "tableRotation",
        "resetCameraDefault3d",
        "resetCameraLastSave3d",
        "hoverClosest3d",
        "zoomInGeo",
        "zoomOutGeo",
        "resetGeo",
        "hoverClosestGeo",
        "hoverClosestGl2d",
        "hoverClosestPie",
        "toggleHover",
        "resetViews",
        "toggleSpikelines",
        "resetViewMapbox",
    ],
}

""" Initial plot layout before data is present """
initial_figure = {
    "layout": {
        "dragmode": "pan",
        "plot_bgcolor": BACKGROUND_NORMAL,
        "paper_bgcolor": BACKGROUND_NORMAL,
        "xaxis": {
            "showgrid": False,
            "zeroline": False,
            "showticklabels": False,
        },
        "yaxis": {"showgrid": False, "zeroline": False, "showticklabels": False},
        "font": {
            "family": "SFMono-Regular,Menlo,Monaco,Consolas,Liberation Mono,Courier,monospace",
            "color": "#e1e1e1",
        },
        "annotations": [
            {
                "text": "Loading...",
                "align": "center",
                "showarrow": False,
                "xref": "paper",
                "yref": "paper",
                "xanchor": "center",
                "yanchor": "center",
                "x": 0.5,
                "y": 0.5,
                "yshift": 0,
                "font": {
                    "size": 16,
                },
            }
        ],
    },
}


def make_layout(
    title=None,
    subtitle=None,
    legend=False,
    date_range_slider=False,
    x_title=None,
    y_title=None,
    y2=False,
    y2_title=None,
    x_hidden=False,
    y_hidden=False,
    source_hidden=True,
    override={},
):
    """ Builds Plotly layout that complies with the Beneath theme """

    layout = {}

    title_height = 24 if title else 0
    subtitle_height = 16 if subtitle else 0

    layout["margin"] = {
        "l": 20 if y_hidden else 60 if y_title else 35,
        "r": 20 if not y2 else 60 if y2_title else 35,
        "t": 20 + (8 if title or subtitle else 0) + title_height + subtitle_height,
        "b": 20 if x_hidden else 50 if x_title else 30,
        "autoexpand": True,
    }

    # title assigned for metadata, but hidden as it is rendered by an annotation (see later)
    layout["title"] = title
    layout["titlefont"] = {
        "size": 1,
        "color": BACKGROUND_NORMAL,
    }

    layout["font"] = {
        "family": "SFMono-Regular,Menlo,Monaco,Consolas,Liberation Mono,Courier,monospace",
        "color": "#e1e1e1",
    }

    layout["plot_bgcolor"] = BACKGROUND_NORMAL
    layout["paper_bgcolor"] = BACKGROUND_NORMAL

    layout["dragmode"] = "select"

    layout["colorway"] = COLORWAY

    layout["xaxis"] = {
        "gridcolor": "rgb(44, 53, 72)",
        "zerolinecolor": "#e1e1e1",
    }

    layout["yaxis"] = {
        "gridcolor": "rgb(44, 53, 72)",
        "zerolinecolor": "#e1e1e1",
    }

    if x_title:
        layout["xaxis"]["title"] = x_title
        layout["xaxis"]["titlefont"] = {"size": 12}

    if y_title:
        layout["yaxis"]["title"] = y_title
        layout["xaxis"]["titlefont"] = {"size": 12}

    if y2:
        layout["yaxis2"] = {
            "title": y2_title,
            "side": "right",
            "overlaying": "y",
            "showgrid": False,
            "zeroline": False,
        }

    if date_range_slider:
        layout["xaxis"]["type"] = "date"
        layout["xaxis"]["rangeslider"] = {
            "visible": True,
            "bgcolor": BACKGROUND_LIGHT,
            "thickness": 0.1,
        }

    if legend:
        layout["legend"] = {
            "orientation": "h",
            "yanchor": "top",
            "x": 0,
            "y": -0.1
            - (0.16 if date_range_slider else 0.0)
            - (0.16 if x_title else 0.0),
        }

    if x_hidden:
        layout["xaxis"] = {
            "showgrid": False,
            "zeroline": False,
            "showticklabels": False,
        }

    if y_hidden:
        layout["yaxis"] = {
            "showgrid": False,
            "zeroline": False,
            "showticklabels": False,
        }

    layout["shapes"] = []
    if "shapes" in override:
        layout["shapes"] = override["shapes"]
        del override["shapes"]

    layout["annotations"] = []
    if "annotations" in override:
        layout["annotations"] = override["annotations"]
        del override["annotations"]

    # watermark
    if not source_hidden:
        layout["annotations"].append(
            {
                "text": '<a style="color: rgb(100, 130, 180);" href="https://beneath.dev/">beneath.dev</a>',
                "align": "right",
                "showarrow": False,
                "xref": "paper",
                "yref": "paper",
                "xanchor": "right",
                "yanchor": "bottom",
                "x": 1.0,
                "y": 1.0,
                "height": title_height + subtitle_height,
                "yshift": 13,
            }
        )

    # subtitle
    if subtitle:
        layout["annotations"].append(
            {
                "text": subtitle,
                "align": "left",
                "showarrow": False,
                "xref": "paper",
                "yref": "paper",
                "xanchor": "left",
                "yanchor": "bottom",
                "x": 0.0,
                "y": 1.0,
                "yshift": 13,
                "height": 16,
                "font": {
                    "size": 12,
                },
            }
        )

    # title
    if title:
        layout["annotations"].append(
            {
                "text": "<b>{}</b>".format(title),
                "align": "left",
                "showarrow": False,
                "xref": "paper",
                "yref": "paper",
                "xanchor": "left",
                "yanchor": "bottom",
                "x": 0.0,
                "y": 1.0,
                "yshift": 13 + subtitle_height,
                "height": 24,
                "font": {
                    "size": 18,
                },
            }
        )

    # totally hacky pseudo-deep merge of `override` into `layout`
    for k1, v1 in override.items():
        if isinstance(v1, dict):
            if k1 not in layout:
                layout[k1] = {}
            for k2, v2 in v1.items():
                if isinstance(v2, dict):
                    if k2 not in layout[k1]:
                        layout[k1][k2] = {}
                    for k3, v3 in v2.items():
                        layout[k1][k2][k3] = v3
                else:
                    layout[k1][k2] = v2
        else:
            layout[k1] = v1

    return layout


def make_double_plot(
    data1, data2, title, subtitles, subtitle=None, vertical=False, override={}
):
    rows = 2 if vertical else 1
    cols = 1 if vertical else 2

    fig = tools.make_subplots(
        rows=rows, cols=cols, vertical_spacing=0.25, horizontal_spacing=0.1
    )

    # trace2.marker.color = COLORWAY[0]

    for trace in data1:
        fig.append_trace(trace, 1, 1)

    for trace in data2:
        fig.append_trace(trace, rows, cols)

    layout = make_layout(
        title=title,
        subtitle=subtitle,
        override={
            "showlegend": False,
            "xaxis2": {
                "gridcolor": "rgb(44, 53, 72)",
            },
            "yaxis2": {
                "showgrid": True,
                "gridcolor": "rgb(44, 53, 72)",
                "zerolinecolor": "#e1e1e1",
            },
            "margin": {"t": 86 + (16 if subtitle else 0), "l": 70, "b": 40},
        },
    )

    # shift titles upwards to make space for subplot titles
    for annotation in layout["annotations"]:
        annotation["yshift"] += 30

    # add subplot titles
    layout["annotations"].append(
        {
            "text": subtitles[0],
            "align": "left",
            "height": 16,
            "x": 0.0,
            "y": 1.0,
            "yshift": 6,
            "xanchor": "left",
            "yanchor": "bottom",
            "xref": "paper",
            "yref": "paper",
            "showarrow": False,
            "font": {"size": 14},
        }
    )
    layout["annotations"].append(
        {
            "text": subtitles[1],
            "align": "left",
            "height": 16,
            "x": 0.0 if vertical else 0.5,
            "y": 0.5 if vertical else 1.0,
            "yshift": 6,
            "xshift": 50,
            "xanchor": "left",
            "yanchor": "bottom",
            "xref": "paper",
            "yref": "paper",
            "showarrow": False,
            "font": {"size": 14},
        }
    )

    fig["layout"].update(layout)
    fig["layout"].annotations = layout["annotations"]
    fig["layout"].update(override)

    return fig


# Saves plot as image to local file, updates plot on plot.ly account, shows plot in Jupyter
def plot_and_save(fig, name):
    # pio.write_image(
    #     fig, "images/{}.png".format(name), width=700, height=432, scale=2, engine="auto"
    # )
    return plt.iplot(
        fig,
        filename=name,
        config={
            "showLink": False,
            "displayModeBar": False,
        },
    )
