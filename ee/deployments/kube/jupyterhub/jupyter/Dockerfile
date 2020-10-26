FROM jupyter/scipy-notebook:latest

RUN pip install bqplot voila dash dash-renderer dash-html-components dash-core-components dash-table dash-cytoscape

RUN pip install jupyterlab_iframe && \
  jupyter labextension install jupyterlab_iframe && \
  jupyter serverextension enable --py jupyterlab_iframe

RUN jupyter labextension install \
  @mflevine/jupyterlab_html \
  jupyterlab-drawio \
  jupyterlab-chart-editor \
  @jupyterlab/plotly-extension

# todo: switch to package install when becomes available
RUN git clone https://github.com/plotly/jupyterlab-dash && \
  cd jupyterlab-dash && \
  npm install && \
  npm run build && \
  jupyter labextension link .

RUN pip install nbserverproxy && jupyter serverextension enable --py nbserverproxy

# pip install --upgrade google-cloud-bigquery

# - exposing service: https://bokeh.pydata.org/en/latest/docs/user_guide/notebook.html#jupyterhub 
# - cool: https://github.com/plotly/plotlylab/blob/master/notebooks/4.%20Chart%20Editor.ipynb

# !pip install psycopg2-binary
# !pip install google-cloud-bigquery
# !pip install squarify